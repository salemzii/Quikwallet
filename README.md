# Backend Developer Task – Wallet API

live project on https://quikwallet.herokuapp.com/

# endpoints 

## . Register /api/v1/auth/register
    method = POST
    fields = {"username":"", "email":"", "password":""} 
    response = {
      		"username": "",
			"email":    "",
			"password": "",
			"your_wallet_id":"",
            }
 On create of player an AfterCreate hook is included to 
 automatically create a wallet,
 ```Go 
	 func (player *Player) AfterCreate(tx *gorm.DB) (err error) {

		w := Wallet{
			Balance: decimal.NewFromFloat(0.00),
		}
		if err := tx.Create(&w).Error; err != nil {
			return err
		}

		return nil
	}
 ```
        
        
  ## . Login /api/v1/auth/login
    method = POST
    fields = {"username":"", "password":""}
    response = {
      "message":"user successfully authenticated"
      }
 
 
 ## .Get wallet balance private/api/v1/wallets/wallet_id/balance
      method = GET
      response = {
        "balance":"",
        "from_cache":bool
        }
 
 ## . Create Wallet private/api/v1/wallets/create
 
      method = POST
      fields = {}
      response = {
              "id":      int,
              "balance": decimal.Decimal,
              "created": time.Time,
              }
              
   ## . Credit wallet private/api/v1/wallets/:wallet_id/credit
   
        method = POST
        fields = {"amount":""}
        response = {
          "balance": decimal.Decimal
          }
          
   ## . Debit wallet private/api/v1/wallets/:wallet_id/debit
   
        method = POST
        fields = {"amount":""}
        response = {
          "balance": decimal.Decimal
          }
          
   ## . Logout /api/v1/auth/logout
        method = GET
        response = {
          "message": "Successfully logged out"
          }
        
