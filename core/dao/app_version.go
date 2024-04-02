package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

func AddAppVersion(ctx context.Context, appVer *model.AppVersion) error {
	statement := `INSERT INTO app_version(version, min_version, description, url, platform, size, lang, created_at, updated_at) VALUES(:version, :min_version, :description, :url, :platform, :size, :lang, :created_at, :updated_at)`
	_, err := DB.NamedExecContext(ctx, statement, appVer)
	return err
}

func UpdateAppVersion(ctx context.Context, appVer *model.AppVersion) error {
	statement := `UPDATE app_version set min_version = ?, description = ?, url = ?, size = ?, updated_at = now() where version = ? and platform = ? and lang = ?`
	_, err := DB.ExecContext(ctx, statement, appVer.MinVersion, appVer.Description, appVer.Url, appVer.Size, appVer.Version, appVer.Platform, appVer.Lang)
	return err
}

func GetAppVersion(ctx context.Context, version string, platform string, lang model.Language) (*model.AppVersion, error) {
	query := `SELECT * from app_version where version = ? and platform = ? and lang = ?`
	var out model.AppVersion
	err := DB.GetContext(ctx, &out, query, version, platform, lang)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func GetLatestAppVersion(ctx context.Context, platform string, lang model.Language) (model.AppVersion, error) {
	query := `SELECT * from app_version where platform = ? and lang = ? order by created_at desc limit 1`
	var out model.AppVersion
	err := DB.GetContext(ctx, &out, query, platform, lang)
	if errors.Is(err, sql.ErrNoRows) {
		return model.AppVersion{}, nil
	}

	if err != nil {
		return model.AppVersion{}, err
	}

	return out, nil
}
