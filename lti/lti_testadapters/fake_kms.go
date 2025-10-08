package lti_testadapters

import (
	"context"

	"github.com/vizdos-enterprises/go-lti/internal/adapters/crypto"
)

type FakeKMSClient struct {
	Pub       []byte
	SignBytes []byte
	Err       error
}

func (f *FakeKMSClient) GetPublicKey(ctx context.Context, in *crypto.GetPublicKeyInput) (*crypto.GetPublicKeyOutput, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	return &crypto.GetPublicKeyOutput{
		PublicKey: f.Pub,
		KeyId:     in.KeyId,
	}, nil
}

func (f *FakeKMSClient) Sign(ctx context.Context, in *crypto.SignInput) (*crypto.SignOutput, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	return &crypto.SignOutput{Signature: f.SignBytes}, nil
}
