package users

import (
	"carpool-btc/internal/app/models"
	"carpool-btc/internal/app/utils"
)

func GetUserWithWallet(wA string) (*models.User, error) {
	queryString := `
		SELECT u.id, u.email_address, u.user_name, u.wallet_address, u.role_id, r.id, r.name, uw.balance
		FROM users u
		LEFT JOIN roles r ON r.id = u.role_id
		LEFT JOIN wallets uw ON uw.user_id = u.id 
		WHERE u.wallet_address = ($1)
	`

	user, err := getUser(queryString, wA)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func getUser(qS string, value string) (*models.User, error) {
	var user models.User
	var role models.Role

	err := utils.DB.QueryRow(qS, value).Scan(
		&user.ID,
		&user.EmailAddress,
		&user.UserName,
		&user.WalletAddress,
		&user.RoleID,
		&role.ID,
		&role.Name,
		&user.Balance,
	)

	if err != nil {
		return nil, err
	}

	user.Role = &role

	return &user, nil
}
