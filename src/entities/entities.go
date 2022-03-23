package entities

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB
var err error

func init() {
	db, err = gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
}

type Player struct {
	gorm.Model
	Username string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

//Generates a password hash for a player's password as storing raw password to db is not ideal
func (p *Player) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// used during login to compare player's login password with the equivalent hash stored in db
func (p *Player) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

type Wallet struct {
	gorm.Model
	//Player      Player          `json:"player"`
	Balance     decimal.Decimal `json:"balance"`
	LastUpdated time.Time       `json:"lastupdated"`
}

func (w *Wallet) WalletNotBelowZero() bool {
	return w.Balance.IsNegative()
}

type TransactionForm struct {
	Amount decimal.Decimal `json:"amount"`
}

// function for creating a player
func CreatePlayer(c *gin.Context) {
	var player Player
	// parse json data into player instance
	c.BindJSON(&player)

	//Db Transaction to create player
	db.Transaction(func(tx *gorm.DB) error {
		pass, err := player.HashPassword(player.Password)
		if err != nil {
			log.Println(err)
		}
		player.Password = pass
		if err := tx.Create(&player).Error; err != nil {
			return err
		}
		// Return json response after saving player
		c.JSON(200, gin.H{
			"username": player.Username,
			"email":    player.Email,
			"password": player.Password,
		})
		return nil
	})
}

//Query db for a particular player using the player's username
func GetPlayer(username string) (p *Player, err error) {
	var player Player

	db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("username = ?", username).First(&player).Error; err != nil {
			return err
		}
		return nil
	})
	return &player, nil
}
