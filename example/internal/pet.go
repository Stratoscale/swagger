package internal

import (
	"context"

	"github.com/Stratoscale/swagger/example/restapi/operations/pet"
	"github.com/go-openapi/runtime/middleware"
)

type Pet struct{}

func (*Pet) PetCreate(ctx context.Context, params pet.PetCreateParams) middleware.Responder {
	panic("implement me")
}

func (*Pet) PetDelete(ctx context.Context, params pet.PetDeleteParams) middleware.Responder {
	panic("implement me")
}

func (*Pet) PetGet(ctx context.Context, params pet.PetGetParams) middleware.Responder {
	panic("implement me")
}

func (*Pet) PetList(ctx context.Context, params pet.PetListParams) middleware.Responder {
	panic("implement me")
}

func (*Pet) PetUpdate(ctx context.Context, params pet.PetUpdateParams) middleware.Responder {
	panic("implement me")
}
