package templater

import (
	"context"
	"encoding/base64"
	"fmt"

	sodium "github.com/GoKillers/libsodium-go/cryptobox"
	"github.com/google/go-github/v54/github"
	"github.com/sirupsen/logrus"
)

func (t *Templater) SetRepoSecrets(ctx context.Context) error {
	for owner, ownerRepos := range t.config.Repositories {
		for repo := range ownerRepos {
			log := logrus.WithField("repo", fmt.Sprintf("%s/%s", owner, repo))
			keyID, key, err := t.getRepoPublicKey(ctx, owner, repo)
			if err != nil {
				return fmt.Errorf("getting repo public key: %w", err)
			}
			log.WithField("key_id", keyID).Info("fetched repository public key")

			for name, plaintext := range t.config.Secrets {
				ciphertext, exit := sodium.CryptoBoxSeal([]byte(plaintext), key)
				if exit != 0 {
					return fmt.Errorf("encrypting plaintext failed")
				}
				s := &github.EncryptedSecret{
					Name:           name,
					KeyID:          keyID,
					EncryptedValue: base64.StdEncoding.EncodeToString(ciphertext),
				}
				if _, err := t.client.Actions.CreateOrUpdateRepoSecret(ctx, owner, repo, s); err != nil {
					return fmt.Errorf("storing repo secret: %w", err)
				}
				log.WithField("secret", name).Info("created repo secret")
			}
		}
	}

	return nil
}

func (t *Templater) getRepoPublicKey(ctx context.Context, owner, repo string) (keyID string, key []byte, err error) {
	k, _, err := t.client.Actions.GetRepoPublicKey(ctx, owner, repo)
	if err != nil {
		return "", nil, fmt.Errorf("getting repo public key: %w", err)
	}
	decoded, err := base64.StdEncoding.DecodeString(k.GetKey())
	if err != nil {
		return "", nil, fmt.Errorf("decoding public key: %w", err)
	}
	return k.GetKeyID(), decoded, nil
}
