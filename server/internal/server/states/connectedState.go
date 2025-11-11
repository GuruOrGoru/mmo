package states

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/guruorgoru/go-mmo/server/internal/objects"
	"github.com/guruorgoru/go-mmo/server/internal/server"
	"github.com/guruorgoru/go-mmo/server/internal/server/db"
	"github.com/guruorgoru/go-mmo/server/pkg/packets"
	"golang.org/x/crypto/bcrypt"
)

type Connected struct {
	client  server.ClientInterfacer
	logger  *log.Logger
	queries *db.Queries
	dbCtx   context.Context
}

func (c *Connected) Name() string {
	return "Connected"
}

// Inject client and set everything up
func (c *Connected) SetClient(newClient server.ClientInterfacer) {
	c.client = newClient
	logPrefix := fmt.Sprintf("[%v] Client %v: ", c.Name(), c.client.GetId())
	c.logger = log.New(log.Writer(), logPrefix, log.LstdFlags)
	c.queries = newClient.DbTx().Queries
	c.dbCtx = newClient.DbTx().Ctx
}

func (c *Connected) OnEntry() {
	// Newly connected client will get his/her Id first
	c.client.Send(packets.NewId(c.client.GetId()))
	user, err := c.client.DbTx().Queries.CreateUser(c.client.DbTx().Ctx, db.CreateUserParams{
		Username:     "username",
		PasswordHash: "password hash",
	})

	if err != nil {
		c.logger.Printf("Failed to create user: %v", err)
	} else {
		c.logger.Printf("Created user: %v", user)
	}
}

func (c *Connected) OnExit() {
	// TODO
}

func (c *Connected) Handle(senderId uint64, msg packets.Msg) {
	switch msg := msg.(type) {
	case *packets.Packet_LoginRequest:
		c.handleLoginRequest(senderId, msg)

	case *packets.Packet_RegisterRequest:
		c.handleRegisterRequest(senderId, msg)

	}
}

func (c *Connected) handleLoginRequest(senderId uint64, msg *packets.Packet_LoginRequest) {
	if senderId != c.client.GetId() {
		c.logger.Printf("Received login message from another client (Id %d)", senderId)
		return
	}

	username := msg.LoginRequest.Username
	password := msg.LoginRequest.Password

	generalFailedMessage := packets.NewDenyMessage("Incorrect Username Or Password")

	user, err := c.queries.GetUserByUsername(c.dbCtx, strings.ToLower(username))
	if err != nil {
		c.logger.Printf("Error getting user by username %v\n", err)
		c.client.Send(generalFailedMessage)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		c.logger.Printf("User entered wrong password: %v", err)
		c.client.Send(generalFailedMessage)
		return
	}
	c.logger.Printf("user %v logged in succesfully", username)
	c.client.Send(packets.NewOkMessage())
	c.client.SetState(&InGame{
		player: &objects.Player{
			Name: username,
		},
	})
}

func (c *Connected) handleRegisterRequest(senderId uint64, msg *packets.Packet_RegisterRequest) {
	if senderId != c.client.GetId() {
		c.logger.Printf("Received register message from another client (Id %d)", senderId)
		return

	}

	username := strings.ToLower(msg.RegisterRequest.Username)
	password := msg.RegisterRequest.Password
	err := validateUsername(msg.RegisterRequest.Username)
	if err != nil {
		reason := fmt.Sprintf("Invalid username: %v", err)
		c.logger.Println(reason)
		c.client.Send(packets.NewDenyMessage(reason))
		return
	}

	_, err = c.queries.GetUserByUsername(c.dbCtx, username)
	if err == nil {
		c.logger.Printf("User with username %v already exists", username)
		c.client.Send(packets.NewDenyMessage("User already exists"))
	}
	genericFailMessage := packets.NewDenyMessage("Error registering user (internal server error) - please try again later")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.logger.Printf("Failed to hash password for %v : %v", username, err)
		c.client.Send(genericFailMessage)
		return
	}

	_, err = c.queries.CreateUser(c.dbCtx, db.CreateUserParams{
		Username: username,
		PasswordHash: string(hashedPassword),
	})
	if err != nil {
		c.logger.Printf("Error creating user %v: %v", username, err)
		c.client.Send(genericFailMessage)
		return
	}

	c.client.Send(packets.NewOkMessage())

	c.logger.Printf("User %v registered successfully", username)
}

func validateUsername(username string) error {
	if len(username) <= 0 {
		return errors.New("empty")
	}

	if len(username) > 20 {
		return errors.New("bruh too long")
	}

	if username != strings.TrimSpace(username) {
		return errors.New("leading or trailing whitespaces")
	}
	
	return nil
}
