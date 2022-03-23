package main

import (
	"log"
	"strconv"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/salemzii/Quikwallet/src/auth"
	"github.com/salemzii/Quikwallet/src/cache"
	entity "github.com/salemzii/Quikwallet/src/entities"

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
	db.AutoMigrate(&entity.Player{})
	db.AutoMigrate(&entity.Wallet{})
}

func main() {

	//create http router
	router := gin.Default()

	router.Use(sessions.Sessions("mysession", sessions.NewCookieStore(auth.Secret)))
	router.Use(JSONMiddleware())
	router.GET("/", welcome)

	//Authentication
	router.POST("/api/v1/auth/register", entity.CreatePlayer)
	router.POST("/api/v1/auth/login", auth.LoginFunc)
	router.GET("/api/v1/auth/logout", auth.Logout)

	// Private group, require authentication to access any wallet resources
	private := router.Group("/private")
	private.Use(auth.AuthRequired)
	{
		private.GET("/api/v1/wallets/:wallet_id/balance", getWalletBalance)
		private.POST("/api/v1/wallets/:wallet_id/credit", creditWallet)
		private.POST("/api/v1/wallets/:wallet_id/debit", debitWallet)
		router.POST("/api/v1/wallets/create", createWallet)
	}

	router.Run()
}
func JSONMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Next()
	}
}

func welcome(c *gin.Context) {

	c.JSON(200, gin.H{
		"message":       "Hello welcome to QuikWallet api",
		"register":      "/api/v1/auth/register",
		"login":         "/api/v1/auth/login",
		"balance":       "private/api/v1/wallets/wallet_id/balance",
		"credit":        "private/api/v1/wallets/:wallet_id/credit",
		"debit":         "private/api/v1/wallets/:wallet_id/debit",
		"create_wallet": "private/api/v1/wallets/create",
		"logout":        "/api/v1/auth/logout",
	})
}

// function for getting the balance of a specific wallet
func getWalletBalance(c *gin.Context) {
	// get wallet id from url parameter
	wallet_id, err := strconv.Atoi(c.Param("wallet_id"))
	if err != nil {
		log.Fatal(err)
	}

	// try to get the wallet balance from cache if it's been requested previously.
	w, err := cache.GetWalletBalanceInCache(wallet_id)

	if err != nil {
		log.Println(err)
	} else {
		c.JSON(200, gin.H{
			"balance":    w.Balance,
			"from_cache": true,
		})
		return
	}

	var wallet entity.Wallet
	//Query db for a particular wallet using it's wallet id
	db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", wallet_id).First(&wallet).Error; err != nil {
			c.JSON(400, gin.H{
				"error": "wallet with id " + strconv.Itoa(wallet_id) + " not found!",
			})

		} else {
			// save wallet balance to cache once it is found NB: all keys are deleted from cache after one hour.
			if err := cache.SetWalletBalanceInCache(wallet_id, wallet.Balance); err != nil {
				log.Println("unable to set value \n", err)
			}
			// return json response with wallet balance
			c.JSON(200, gin.H{
				"balance":    wallet.Balance,
				"from_cache": false,
			})
		}
		return nil
	})
}

//function to credit a wallet
func creditWallet(c *gin.Context) {
	// get wallet id from url parameter
	wallet_id, err := strconv.Atoi(c.Param("wallet_id"))

	if err != nil {
		log.Fatal(err)
	}
	// declare two instances of wallet, one for query db, the other for recieving credit amount
	var wallet entity.Wallet
	var creditamount entity.TransactionForm

	//Query db for a particular wallet using it's wallet id
	db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", wallet_id).First(&wallet).Error; err != nil {
			c.JSON(400, gin.H{
				"error": "wallet with id " + strconv.Itoa(wallet_id) + " not found!",
			})

		} else {
			// parse json data into wallet instance
			c.BindJSON(&creditamount)

			//Check if amount to be credited is a positive number
			if creditamount.Amount.IsPositive() {

				// update db balance with the new credit amount
				tx.Model(&wallet).Update("balance", wallet.Balance.Add(creditamount.Amount))

				//update cache to set balance for that key to it's new value
				if err := cache.SetWalletBalanceInCache(wallet_id, wallet.Balance); err != nil {
					log.Println("unable to set value \n", err)
				}
				c.JSON(200, gin.H{
					"balance": wallet.Balance,
				})
			} else {
				//return status 400 if amount to be credited is negative.
				c.JSON(400, gin.H{
					"error": "Cannot use negative value " + creditamount.Amount.String() + " for operation",
				})
			}
		}
		return nil
	})
}

//function to credit a wallet
func debitWallet(c *gin.Context) {
	// get wallet id from url parameter
	wallet_id, err := strconv.Atoi(c.Param("wallet_id"))

	if err != nil {
		log.Fatal(err)
	}
	// declare two instances of wallet, one for query db, the other for recieving dedit amount
	var wallet entity.Wallet
	var debitamount entity.TransactionForm
	//Query db for a particular wallet using it's wallet id
	db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", wallet_id).First(&wallet).Error; err != nil {
			c.JSON(400, gin.H{
				"error": "wallet with id " + strconv.Itoa(wallet_id) + " not found!",
			})
		} else {
			// parse json data into wallet instance
			c.BindJSON(&debitamount)

			//Check if amount to be dedited is a positive number
			if debitamount.Amount.IsPositive() {
				if !(debitamount.Amount.GreaterThan(wallet.Balance)) && !(wallet.WalletNotBelowZero()) {
					// update db balance with the new dedit amount
					tx.Model(&wallet).Update("balance", wallet.Balance.Sub(debitamount.Amount))

					//update cache to set balance for that key to it's new value
					if err := cache.SetWalletBalanceInCache(wallet_id, wallet.Balance); err != nil {
						log.Println("unable to set value \n", err)
					}
					c.JSON(200, gin.H{
						"balance": wallet.Balance,
					})
				} else {
					// return status 400 if amount to be debitted is greater than wallet balance
					c.JSON(400, gin.H{
						"error": "Insufficient Balance for operation",
					})
				}
			} else {
				//return status 400 if amount to be dedited is negative.
				c.JSON(400, gin.H{
					"error": "Cannot use negative value " + debitamount.Amount.String() + " for operation",
				})
			}
		}
		return nil
	})
}

// function for creating a wallet
func createWallet(c *gin.Context) {
	var wallet entity.Wallet
	// parse json data into wallet instance
	c.BindJSON(&wallet)

	//Db Transaction to create wallet
	db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&wallet).Error; err != nil {
			return err
		}

		// Return json response after saving wallet
		c.JSON(200, gin.H{
			"id":      wallet.ID,
			"balance": wallet.Balance,
			"created": wallet.CreatedAt,
		})
		return nil
	})
}
