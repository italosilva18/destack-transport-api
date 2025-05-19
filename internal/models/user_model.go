package models

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User representa um usuário no sistema
type User struct {
	BaseModel
	Name     string `json:"name" gorm:"not null"`
	Username string `json:"username" gorm:"uniqueIndex;not null"`
	Email    string `json:"email" gorm:"uniqueIndex;not null"`
	Password string `json:"-" gorm:"not null"` // O "-" indica que este campo não será incluído em JSON
	Role     string `json:"role" gorm:"default:user;not null"`
	Active   bool   `json:"active" gorm:"default:true;not null"`
}

// TableName define o nome da tabela no banco de dados
func (User) TableName() string {
	return "users"
}

// BeforeSave é chamado antes de salvar o usuário no banco
func (u *User) BeforeSave(tx *gorm.DB) error {
	// Apenas hash a senha se ela for alterada
	if u.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashedPassword)
	}
	return nil
}

// CheckPassword verifica se a senha fornecida corresponde à senha armazenada
func (u *User) CheckPassword(password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return errors.New("senha incorreta")
	}
	return nil
}
