package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/salemzii/Quikwallet/src/cache"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Player struct {
	gorm.Model
	Username string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Wallet struct {
	gorm.Model
	//Player      Player          `json:"player"`
	Balance     decimal.Decimal `json:"balance`
	LastUpdated time.Time       `json:"lastupdated"`
}

func (w *Wallet) WalletNotBelowZero() bool {
	return w.Balance.IsNegative()
}

var db *gorm.DB
var err error

func init() {
	db, err = gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&Player{})
	db.AutoMigrate(&Wallet{})
}

func main() {

	//create http router
	router := gin.Default()

	router.GET("/", welcome)
	router.GET("/api/v1/wallets/:wallet_id/balance", getWalletBalance)
	router.POST("/api/v1/wallets/:wallet_id/credit", creditWallet)
	router.POST("/api/v1/wallets/:wallet_id/debit", debitWallet)

	router.POST("/api/v1/wallets/create", createWallet)

	router.Run()
}

func welcome(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "welcome",
	})
}

func createWallet(c *gin.Context) {
	var wallet Wallet
	c.BindJSON(&wallet)
	db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&wallet).Error; err != nil {
			return err
		}
		c.JSON(200, gin.H{
			"id":      wallet.ID,
			"balance": wallet.Balance,
			"created": wallet.CreatedAt,
		})
		return nil
	})
}
func getWalletBalance(c *gin.Context) {
	wallet_id, err := strconv.Atoi(c.Param("wallet_id"))
	if err != nil {
		log.Fatal(err)
	}

	w, err := cache.GetWalletBalanceInCache(wallet_id)

	if err != nil {
		log.Fatal(err)
	} else {
		c.JSON(200, gin.H{
			"balance": w.Balance,
		})
	}

	var wallet Wallet
	db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", wallet_id).First(&wallet).Error; err != nil {
			c.JSON(400, gin.H{
				"error": "wallet with id " + strconv.Itoa(wallet_id) + " not found!",
			})

		} else {
			if err := cache.SetWalletBalanceInCache(wallet_id, wallet.Balance); err != nil {
				log.Println("unable to set value \n", err)
			}
			c.JSON(200, gin.H{
				"balance": wallet.Balance,
			})
		}
		return nil
	})
}

func creditWallet(c *gin.Context) {
	wallet_id, err := strconv.Atoi(c.Param("wallet_id"))

	if err != nil {
		log.Fatal(err)
	}
	var wallet Wallet
	var wallet2 Wallet
	db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", wallet_id).First(&wallet).Error; err != nil {
			c.JSON(400, gin.H{
				"error": "wallet with id " + strconv.Itoa(wallet_id) + " not found!",
			})

		} else {
			c.BindJSON(&wallet2)
			if wallet2.Balance.IsPositive() {

				tx.Model(&wallet).Update("balance", wallet.Balance.Add(wallet2.Balance))
				c.JSON(200, gin.H{
					"balance": wallet.Balance,
				})
			} else {
				c.JSON(400, gin.H{
					"error": "Cannot use negative value " + wallet2.Balance.String() + " for operation",
				})
			}
		}
		return nil
	})
}

func debitWallet(c *gin.Context) {
	wallet_id, err := strconv.Atoi(c.Param("wallet_id"))

	if err != nil {
		log.Fatal(err)
	}
	var wallet Wallet
	var wallet2 Wallet
	db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", wallet_id).First(&wallet).Error; err != nil {
			c.JSON(400, gin.H{
				"error": "wallet with id " + strconv.Itoa(wallet_id) + " not found!",
			})
		} else {
			c.BindJSON(&wallet2)
			if wallet2.Balance.IsPositive() {
				if !(wallet2.Balance.GreaterThan(wallet.Balance)) && !(wallet.WalletNotBelowZero()) {
					tx.Model(&wallet).Update("balance", wallet.Balance.Sub(wallet2.Balance))
					fmt.Println(wallet.Balance)
					c.JSON(200, gin.H{
						"balance": wallet.Balance,
					})
				} else {
					c.JSON(400, gin.H{
						"error": "Insufficient Balance for operation",
					})
				}
			} else {
				c.JSON(400, gin.H{
					"error": "Cannot use negative value " + wallet2.Balance.String() + " for operation",
				})
			}
		}
		return nil
	})
}
