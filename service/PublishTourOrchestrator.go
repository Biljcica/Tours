package service

import (
	"database-example/repo"
    "database-example/model"
    "context"
    "log"
    "time"
	"github.com/andjelavukosav/Docker/common/saga/messaging"
	"github.com/andjelavukosav/Docker/common/saga/publish_tour"
)

type PublishTourOrchestrator struct {
	commandPublisher messaging.Publisher
	replySubscriber  messaging.Subscriber
	TourRepo         *repo.TourRepository
}


func NewPublishTourOrchestrator(publisher messaging.Publisher, subscriber messaging.Subscriber, tourRepo *repo.TourRepository) (*PublishTourOrchestrator, error) {
	o := &PublishTourOrchestrator{
		commandPublisher: publisher,
		replySubscriber:  subscriber,
		TourRepo:         tourRepo,
	}
	err := o.replySubscriber.Subscribe(o.handle)
	if err != nil {
		return nil, err
	}
	log.Println("PublishTourOrchestrator subscribed successfully")
	return o, nil
}


func (o *PublishTourOrchestrator) Start(tour publish_tour.TourDetails) error {
	event := &publish_tour.PublishTourCommand{
		Tour: tour,
		Type: publish_tour.PublishTour, // inicijalna komanda
	}
	log.Printf("[SAGA] Attempting to publish command for tour %s: %+v", tour.ID, event)

	return o.commandPublisher.Publish(event)
}

func (o *PublishTourOrchestrator) handle(reply *publish_tour.PublishTourReply) {
	command := publish_tour.PublishTourCommand{Tour: reply.Tour}
	command.Type = o.nextCommandType(reply.Type)

	if command.Type != publish_tour.UnknownCommand {
		_ = o.commandPublisher.Publish(&command)
	}

	// Update status u bazi
	ctx := context.Background()
	switch reply.Type {
	case publish_tour.TourPublishedSuccessfully:
		_, err := o.TourRepo.ChangeTourStatus(ctx, reply.Tour.ID, map[string]interface{}{
			"status":      model.Published,
			"publishedAt": time.Now(),
		})
		if err != nil {
			log.Printf("Failed to update tour status to Published: %v", err)
		} else {
			log.Printf("Tour %s successfully published", reply.Tour.ID)
		}
	case publish_tour.TourPublishFailed:
		_, err := o.TourRepo.ChangeTourStatus(ctx, reply.Tour.ID, map[string]interface{}{
			"status": model.Draft,
		})
		if err != nil {
			log.Printf("Failed to revert tour status to Draft: %v", err)
		} else {
			log.Printf("Tour %s publishing failed, reverted to Draft", reply.Tour.ID)
		}
	}
}

func (o *PublishTourOrchestrator) nextCommandType(reply publish_tour.PublishTourReplyType) publish_tour.PublishTourCommandType {
	switch reply {
	case publish_tour.TourPublishedSuccessfully:
		// ovde možeš da pošalješ npr. notifikaciju turistima
		return publish_tour.UnknownCommand
	case publish_tour.TourPublishFailed:
		return publish_tour.CancelTour
	case publish_tour.TourCancelled:
		return publish_tour.UnknownCommand
	default:
		return publish_tour.UnknownCommand
	}
}
