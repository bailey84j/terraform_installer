package password

import (
	"crypto/rand"
	"math/big"
	"os"
	"path/filepath"

	"github.com/bailey84j/terraform_installer/pkg/asset"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

var (
	// tfePasswordPath is the path where tfe user password is stored.
	tfePasswordPath = filepath.Join("auth", "tfe-password")
)

// TFEPassword is the asset for the tfe user password
type TFEPassword struct {
	Password     string
	PasswordHash []byte
	File         *asset.File
}

var _ asset.WritableAsset = (*TFEPassword)(nil)

// Dependencies returns no dependencies.
func (a *TFEPassword) Dependencies() []asset.Asset {
	return []asset.Asset{}
}

// Generate the tfe password
func (a *TFEPassword) Generate(asset.Parents) error {
	logrus.Debugf("Trace Me - Password - Generate...")
	err := a.generateRandomPasswordHash(23)
	if err != nil {
		return err
	}
	return nil
}

// generateRandomPasswordHash generates a hash of a random ASCII password
// 5char-5char-5char-5char
func (a *TFEPassword) generateRandomPasswordHash(length int) error {
	const (
		lowerLetters = "abcdefghijkmnopqrstuvwxyz"
		upperLetters = "ABCDEFGHIJKLMNPQRSTUVWXYZ"
		digits       = "23456789"
		all          = lowerLetters + upperLetters + digits
	)
	var password string
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(all))))
		if err != nil {
			return err
		}
		newchar := string(all[n.Int64()])
		if password == "" {
			password = newchar
		}
		if i < length-1 {
			n, err = rand.Int(rand.Reader, big.NewInt(int64(len(password)+1)))
			if err != nil {
				return err
			}
			j := n.Int64()
			password = password[0:j] + newchar + password[j:]
		}
	}
	pw := []rune(password)
	for _, replace := range []int{5, 11, 17} {
		pw[replace] = '-'
	}
	if a.Password == "" {
		a.Password = string(pw)
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(a.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	a.PasswordHash = bytes
	logrus.Debugf("Trace Me - Password - Create File")
	a.File = &asset.File{
		Filename: tfePasswordPath,
		Data:     []byte(a.Password),
	}

	return nil
}

// Name returns the human-friendly name of the asset.
func (a *TFEPassword) Name() string {
	return "TFE Password"
}

// Files returns the password file.
func (a *TFEPassword) Files() []*asset.File {
	if a.File != nil {
		return []*asset.File{a.File}
	}
	return []*asset.File{}
}

// Load loads a predefined hash only, if one is supplied
func (a *TFEPassword) Load(f asset.FileFetcher) (found bool, err error) {
	hashFilePath := filepath.Join("tls", "tfe-password.hash")
	hashFile, err := f.FetchByName(hashFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	a.PasswordHash = hashFile.Data
	// Assisted-service expects to always see a password file, so generate an
	// empty one
	a.File = &asset.File{
		Filename: tfePasswordPath,
	}
	return true, nil
}
