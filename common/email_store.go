package common

import "database/sql"
import "github.com/Azareal/Gosora/query_gen"

var Emails EmailStore

type EmailStore interface {
	GetEmailsByUser(user *User) (emails []Email, err error)
	VerifyEmail(email string) error
}

type DefaultEmailStore struct {
	getEmailsByUser *sql.Stmt
	verifyEmail     *sql.Stmt
}

func NewDefaultEmailStore(acc *qgen.Accumulator) (*DefaultEmailStore, error) {
	return &DefaultEmailStore{
		getEmailsByUser: acc.Select("emails").Columns("email, validated, token").Where("uid = ?").Prepare(),

		// Need to fix this: Empty string isn't working, it gets set to 1 instead x.x -- Has this been fixed?
		verifyEmail: acc.Update("emails").Set("validated = 1, token = ''").Where("email = ?").Prepare(),
	}, acc.FirstError()
}

func (store *DefaultEmailStore) GetEmailsByUser(user *User) (emails []Email, err error) {
	email := Email{UserID: user.ID}
	rows, err := store.getEmailsByUser.Query(user.ID)
	if err != nil {
		return emails, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&email.Email, &email.Validated, &email.Token)
		if err != nil {
			return emails, err
		}

		if email.Email == user.Email {
			email.Primary = true
		}
		emails = append(emails, email)
	}
	return emails, rows.Err()
}

func (store *DefaultEmailStore) VerifyEmail(email string) error {
	_, err := store.verifyEmail.Exec(email)
	return err
}
