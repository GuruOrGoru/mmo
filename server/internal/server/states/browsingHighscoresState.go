package states

import (
	"context"
	"fmt"
	"log"

	"github.com/guruorgoru/go-mmo/server/internal/server"
	"github.com/guruorgoru/go-mmo/server/internal/server/db"
	"github.com/guruorgoru/go-mmo/server/pkg/packets"
)

type BrowsingHighscores struct {
	clients server.ClientInterfacer
	logger  *log.Logger
	queries *db.Queries
	dbCtx   context.Context
}

func (b *BrowsingHighscores) Name() string {
	return "BrowsingHighScores"
}

func (b *BrowsingHighscores) SetClient(client server.ClientInterfacer) {
	b.clients = client
	logPrefix := fmt.Sprintf("Client %v [%v]", client.GetId(), b.Name())
	b.logger = log.New(log.Writer(), logPrefix, log.LstdFlags)
	b.queries = client.DbTx().Queries
	b.dbCtx = client.DbTx().Ctx
}

func (b *BrowsingHighscores) OnEntry() {
	b.sendTopScorers(50, 0)
}

func (b *BrowsingHighscores) Handle(senderId uint64, msg packets.Msg) {
	switch msg := msg.(type) {
	case *packets.Packet_FinishedBrowsingHighscore:
		b.handleFinishedBrowsingHighscores(msg)
	case *packets.Packet_SearchHighscore:
		b.handleSearchHighscores(senderId, msg)
	}
}

func (b *BrowsingHighscores) handleSearchHighscores(_ uint64, msg *packets.Packet_SearchHighscore) {
	player, err := b.queries.GetPlayerByName(b.dbCtx, msg.SearchHighscore.Name)
	if err != nil {
		b.logger.Printf("error getting user %v %v", msg.SearchHighscore.Name, err)
		b.clients.Send(packets.NewDenyMessage("Error finding the user by the name"))
		return
	}

	playerRank, err := b.queries.GetPlayerRank(b.dbCtx, player.ID)
	if err != nil {
		b.logger.Printf("error getting user rank %v %v", player.ID, err)
		b.clients.Send(packets.NewDenyMessage("Error finding the user rank by the Id"))
		return
	}

	const limit int = 50

	offSet := playerRank - int32(limit)/2
	b.sendTopScorers(limit, int(max(0, offSet)))
}

func (b *BrowsingHighscores) handleFinishedBrowsingHighscores(_ *packets.Packet_FinishedBrowsingHighscore) {
	b.clients.SetState(&Connected{})
}

func (b *BrowsingHighscores) OnExit() {
	// Todo
}

func (b *BrowsingHighscores) sendTopScorers(limit int, offSet int) {
	topScorers, err := b.queries.GetTopScores(b.dbCtx, db.GetTopScoresParams{
		Limit:  int32(limit),
		Offset: int32(offSet),
	})
	if err != nil {
		b.logger.Println("Error fetching top scorers: ", err)
		b.clients.Send(packets.NewDenyMessage("error fetching top scorers, internal server error"))
		return
	}
	hsMessages := make([]*packets.HiscoreMessage, 0, limit)
	for rank, rowScore := range topScorers {
		hsMessages = append(hsMessages, &packets.HiscoreMessage{
			Rank:  uint64(rank) + uint64(offSet) + 1,
			Name:  rowScore.Name,
			Score: uint64(rowScore.BestScore),
		})
	}

	b.clients.Send(packets.NewHighScoreBoard(hsMessages))
}
