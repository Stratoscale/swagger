package pet

import (
	"context"
	"sync"

	"github.com/Stratoscale/swagger/example/models"
	"github.com/Stratoscale/swagger/example/restapi/operations/pet"
	"github.com/go-openapi/runtime/middleware"
)

// A simple in memory CRUD on data
// In real life, this should be using a persistent storage.
type Pet struct {
	data  map[int64]*models.Pet
	count int64
	lock  sync.RWMutex
}

// New returns a new Pet manager
func New() *Pet {
	return &Pet{data: make(map[int64]*models.Pet)}
}

func (p *Pet) PetCreate(ctx context.Context, params pet.PetCreateParams) middleware.Responder {
	p.lock.Lock()
	defer p.lock.Unlock()

	// copy the sent model
	model := *params.Pet

	// set the ID and increment the pets counter
	model.ID = p.count
	p.count++

	// store the created pet
	p.data[params.Pet.ID] = &model

	// copy the stored model before response
	retModel := model

	// return a response
	return pet.NewPetCreateCreated().WithPayload(&retModel)
}

func (p *Pet) PetDelete(ctx context.Context, params pet.PetDeleteParams) middleware.Responder {
	p.lock.Lock()
	defer p.lock.Unlock()

	// check if id exists
	if _, ok := p.data[params.PetID]; !ok {
		return pet.NewPetDeleteNotFound()
	}

	// delete from the storage
	delete(p.data, params.PetID)

	return pet.NewPetDeleteNoContent()
}

func (p *Pet) PetGet(ctx context.Context, params pet.PetGetParams) middleware.Responder {
	p.lock.RLock()
	defer p.lock.RUnlock()

	// check if id exists
	model, ok := p.data[params.PetID]
	if !ok {
		return pet.NewPetGetNotFound()
	}

	// copy the model
	retModel := *model

	return pet.NewPetGetOK().WithPayload(&retModel)
}

func (p *Pet) PetList(ctx context.Context, params pet.PetListParams) middleware.Responder {
	p.lock.RLock()
	defer p.lock.RUnlock()

	// copy the pet list
	data := make([]*models.Pet, 0, len(p.data))
	for _, model := range p.data {
		model := model
		data = append(data, model)
	}

	return pet.NewPetListOK().WithPayload(data)
}

func (p *Pet) PetUpdate(ctx context.Context, params pet.PetUpdateParams) middleware.Responder {
	p.lock.Lock()
	defer p.lock.Unlock()

	// check if id exists
	if _, ok := p.data[params.PetID]; !ok {
		return pet.NewPetDeleteNotFound()
	}

	// copy the model and update it's ID
	model := *params.Pet
	model.ID = params.PetID

	// store the update model
	p.data[params.PetID] = &model

	// copy the updated model
	retModel := model

	// return the updated model
	return pet.NewPetUpdateCreated().WithPayload(&retModel)
}
