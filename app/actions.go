// Copyright 2012 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"fmt"
	"github.com/globocom/config"
	"github.com/globocom/tsuru/api/bind"
	"github.com/globocom/tsuru/db"
	"labix.org/v2/mgo/bson"
	"strconv"
)

// insertApp is an implementation for the action interface.
type insertApp struct{}

// insertApp forward stores the app with "pending" as your state.
func (a *insertApp) forward(app *App) error {
	app.State = "pending"
	return db.Session.Apps().Insert(app)
}

// insertApp backward removes the app from the database.
func (a *insertApp) backward(app *App) {
	db.Session.Apps().Remove(bson.M{"name": app.Name})
}

// createBucketIam is an implementation for the action interface.
type createBucketIam struct{}

// createBucketIam forward creates a bucket and exports
// the related info as environs in the app machine.
func (a *createBucketIam) forward(app *App) error {
	env, err := createBucket(app)
	if err != nil {
		return err
	}
	host, _ := config.GetString("host")
	envVars := []bind.EnvVar{
		{Name: "APPNAME", Value: app.Name},
		{Name: "TSURU_HOST", Value: host},
	}
	variables := map[string]string{
		"ENDPOINT":           env.endpoint,
		"LOCATIONCONSTRAINT": strconv.FormatBool(env.locationConstraint),
		"ACCESS_KEY_ID":      env.AccessKey,
		"SECRET_KEY":         env.SecretKey,
		"BUCKET":             env.bucket,
	}
	for name, value := range variables {
		envVars = append(envVars, bind.EnvVar{
			Name:         fmt.Sprintf("TSURU_S3_%s", name),
			Value:        value,
			InstanceName: s3InstanceName,
		})
	}
	app.SetEnvsToApp(envVars, false, true)
	return err
}

// createBucketIam backward destroys the app bucket.
func (a *createBucketIam) backward(app *App) {
	destroyBucket(app)
}

// deploy is an implementation for the action interface.
type deploy struct{}

// deploy forward deploys the app.
func (a *deploy) forward(app *App) error {
	return app.deploy()
}

// deploy backward does nothing.
func (a *deploy) backward(app *App) {}
