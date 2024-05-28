package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/pkg/random"
	"github.com/golang-module/carbon/v2"
)

const tableNameUser = "users"

func CreateUser(ctx context.Context, user *model.User) error {
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (uuid, username, pass_hash, user_email, wallet_address, role, referrer, referral_code, referrer_user_id, created_at)
			VALUES (:uuid, :username, :pass_hash, :user_email, :wallet_address, :role, :referrer, :referral_code, :referrer_user_id, :created_at);`, tableNameUser,
	), user)

	referralCode := &model.ReferralCode{
		UserId:    user.Username,
		Code:      random.GenerateRandomString(6),
		CreatedAt: user.CreatedAt,
	}

	_, err = tx.NamedExecContext(ctx, `INSERT INTO referral_code(user_id, code, created_at) VALUES (:user_id, :code, :created_at)`, referralCode)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func AddNewReferralCode(ctx context.Context, referralCode *model.ReferralCode) error {
	_, err := DB.NamedExecContext(ctx, `INSERT INTO referral_code(user_id, code, created_at) VALUES (:user_id, :code, :created_at)`, referralCode)
	return err
}

func ResetPassword(ctx context.Context, passHash, username string) error {
	_, err := DB.DB.ExecContext(ctx, fmt.Sprintf(
		`UPDATE %s SET pass_hash = '%s', updated_at = now() WHERE username = '%s'`, tableNameUser, passHash, username))
	return err
}

func GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var out model.User
	if err := DB.QueryRowxContext(ctx, fmt.Sprintf(
		`SELECT * FROM %s WHERE username = ?`, tableNameUser), username,
	).StructScan(&out); err != nil {
		return nil, err
	}

	return &out, nil
}

func GetUserIds(ctx context.Context) ([]string, error) {
	queryStatement := fmt.Sprintf(`SELECT username as user_id FROM %s;`, tableNameUser)
	var out []string
	err := DB.SelectContext(ctx, &out, queryStatement)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRow
		}
		return nil, err
	}
	return out, nil
}

func UpdateUserWalletAddress(ctx context.Context, username, address string) error {
	query := fmt.Sprintf("update %s set wallet_address = ? where username = ?", tableNameUser)
	_, err := DB.ExecContext(ctx, query, address, username)
	return err
}

func GetUserByRefCode(ctx context.Context, refCode string) (*model.User, error) {
	var out model.User
	err := DB.QueryRowxContext(ctx, `SELECT u.* FROM users u join referral_code r on u.username = r.user_id WHERE r.code=? LIMIT 1`, refCode).StructScan(&out)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoRow
		}
		return nil, err
	}
	return &out, nil
}

func GetUserReferCodes(ctx context.Context, userId string) ([]*model.ReferralCode, error) {
	var out []*model.ReferralCode
	err := DB.SelectContext(ctx, &out, `select * from referral_code where user_id = ?`, userId)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func GetReferralCodeProfileByUserId(ctx context.Context, userId string) ([]*model.ReferralCodeProfile, error) {
	var out []*model.ReferralCodeProfile
	query := `select r.code, ifnull(count(u.username),0) as referral_users , ifnull(sum(u.device_count),0) as referral_nodes, ifnull(sum(u.device_online_count),0) as referral_online_nodes, r.created_at from referral_code r left join users u on u.referrer = r.code where  r.user_id = ? group by code`
	err := DB.SelectContext(ctx, &out, query, userId)
	if err != nil {
		return nil, err
	}
	return out, nil

}

func CountReferralUsersByCode(ctx context.Context, code string, option QueryOption) (referralUsers, referralNodes int64, err error) {
	where := ""
	if option.StartTime != "" {
		where += fmt.Sprintf(" AND created_at >= '%s'", option.StartTime)
	}

	if option.EndTime != "" {
		where += fmt.Sprintf(" AND created_at < '%s'", option.EndTime)
	}

	query := `select ifnull(count(username),0) from users where referrer = ?` + where

	row := DB.QueryRowxContext(ctx, query, code)
	err = row.Scan(&referralUsers)

	queryRn := `select ifnull(count(device_id),0) from device_info where user_id in (select username from users where referrer = ?)` + where

	rowRn := DB.QueryRowxContext(ctx, queryRn, code)
	err = rowRn.Scan(&referralNodes)

	return
}

func GetUserReferrerUsersDailyStat(ctx context.Context, code string, option QueryOption) ([]*model.DateValue, error) {
	var out []*model.DateValue

	query := `select date_format(created_at, '%Y-%m-%d') as date, ifnull(count(username),0) as value from users where referrer = ? and created_at >= ? and created_at <= ? group by date`
	err := DB.SelectContext(ctx, &out, query, code, option.StartTime, option.EndTime)
	if err != nil {
		return nil, err
	}

	return appendDataValueList(out, option.StartTime, option.EndTime), nil
}

func GetUserReferrerNodesDailyStat(ctx context.Context, code string, option QueryOption) ([]*model.DateValue, error) {
	var out []*model.DateValue

	query := `select date_format(created_at, '%Y-%m-%d') as date, ifnull(count(device_id),0) as value from device_info where user_id in (select username from users where referrer = ?) and created_at >= ? and created_at <= ? group by date`
	err := DB.SelectContext(ctx, &out, query, code, option.StartTime, option.EndTime)
	if err != nil {
		return nil, err
	}

	return appendDataValueList(out, option.StartTime, option.EndTime), nil
}

func appendDataValueList(dv []*model.DateValue, start, end string) []*model.DateValue {
	startTime, endTime := carbon.Parse(start), carbon.Parse(end)
	deviceInDate := make(map[string]model.DateValue)

	for _, data := range dv {
		deviceInDate[dateKey(carbon.Parse(data.Date).StdTime())] = *data
	}

	var out []*model.DateValue
	for st := startTime.StartOfDay(); st.Lte(endTime.StartOfDay()); st = st.AddDay() {
		val := deviceInDate[dateKey(st.StdTime())]
		out = append(out, &model.DateValue{
			Date:  dateKey(st.StdTime()),
			Value: val.Value,
		})
	}

	return out

}

func GetUserReferralCounter(ctx context.Context, userId string) (*model.ReferralCounter, error) {
	query := `select count(1) as referral_users, sum(device_count) as referral_nodes, sum(reward) as referee_reward from users where referrer_user_id = ?`

	var out model.ReferralCounter
	err := DB.GetContext(ctx, &out, query, userId)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

func GetAllUserReferrerUserId(ctx context.Context) (map[string]string, error) {
	out := make(map[string]string)

	offset := 0
	pageSize := 10000

	for {
		query := `SELECT username, referrer_user_id FROM users WHERE referrer_user_id <> '' LIMIT ? OFFSET ?`
		rows, err := DB.QueryxContext(ctx, query, pageSize, offset)
		if err != nil {
			return nil, err
		}

		var foundRows bool

		for rows.Next() {
			var userId, referrerUserId string
			if err := rows.Scan(&userId, &referrerUserId); err != nil {
				continue
			}

			out[userId] = referrerUserId
			foundRows = true
		}

		rows.Close()

		if !foundRows {
			break
		}

		offset += pageSize
	}

	return out, nil
}
