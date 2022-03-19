package main

import (
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Player struct {
	gorm.Model
	Id       int    `json:"id"`
	Username string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Wallet struct {
	gorm.Model
	Id          int             `json:"id"`
	Balance     decimal.Decimal `json:"balance`
	LastUpdated time.Time       `json:"lastupdated"`
}

func (w *Wallet) GetWalletBalance(walletId int) (bal decimal.Decimal, err error) {
	return w.Balance, nil
}
func (w *Wallet) CreditWalletBalance(walletId int, amount float64) (bal decimal.Decimal, err error) {
	w.Balance = w.Balance.Add(decimal.NewFromFloat(amount))
	return w.Balance, nil
}
func (w *Wallet) DebitWalletBalance(walletId int, amount float64) (bal decimal.Decimal, err error) {
	w.Balance = w.Balance.Sub(decimal.NewFromFloat(amount))
	return w.Balance, nil
}

func init() {}

func main() {

	//create http router
	router := gin.Default()

	router.GET("/", welcome)
	router.GET("/api/v1/wallets/:wallet_id/balance", getWallet)
	router.POST("/api/v1/wallets/:wallet_id/credit", creditWallet)

	router.Run()
}

func welcome(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "welcome",
	})
}
func getWallet(c *gin.Context) {
	wallet_id, err := strconv.Atoi(c.Param("wallet_id"))

	if err != nil {
		log.Fatal(err)
	}
	c.JSON(200, gin.H{
		"balance": wallet_id,
	})
}

func creditWallet(c *gin.Context) {
	wallet_id, err := strconv.Atoi(c.Param("wallet_id"))
	if err != nil {
		log.Fatal(err)
	}
	c.JSON(200, gin.H{
		"balance": wallet_id,
	})
}
