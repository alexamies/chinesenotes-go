// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.


// Package for media metadata
package media

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/alexamies/chinesenotes-go/applog"
)

// A media object structure for media metadata
type MediaMetadata struct {
	ObjectId, TitleZhCn, TitleEn, Author, License string
}

// MediaSearcher looks up media objects.
type MediaSearcher struct {
	database *sql.DB
	findMediaStmt *sql.Stmt
	initialized bool
}

// Open database connection and prepare query
func NewMediaSearcher(database *sql.DB, ctx context.Context) *MediaSearcher {
	applog.Info("media.init Initializing mediameta")
	ms := MediaSearcher{database: database}
	err := ms.InitQuery(ctx)
	if err != nil {
		applog.Errorf("media/NewMediaSearcher: error preparing database statement ", err)
		return &ms
	}
	ms.initialized = true
	return &ms
}

func (ms *MediaSearcher) InitQuery(ctx context.Context) error {
	var err error
	ms.findMediaStmt, err = ms.database.PrepareContext(ctx, 
`SELECT medium_resolution, title_zh_cn, title_en, author, license
FROM illustrations
WHERE medium_resolution = ?
LIMIT 1`)
  if err != nil {
  	return fmt.Errorf("media.initQuery Error preparing fwstmt: %v", err)
  }
  return nil
}

// Looks up media metadata by object ID
func (ms *MediaSearcher) Initialized() bool {
	return ms.initialized
}

// Looks up media metadata by object ID
func (ms *MediaSearcher) FindMedia(objectId string, ctx context.Context) (*MediaMetadata, error) {
	applog.Info("FindMedia: objectId (len) ", objectId, len(objectId))
	mediaMeta := MediaMetadata{}
	results, err := ms.findMediaStmt.QueryContext(ctx, objectId)
  if err != nil {
  	return nil, fmt.Errorf("media.FindMedia Error executing query: %v", err)
  }
	results.Next()
	var medium, titleZhCn, titleEn, author, license sql.NullString
	results.Scan(&medium, &titleZhCn, &titleEn, &author, &license)
	results.Close()
	if medium.Valid {
		mediaMeta.ObjectId = medium.String
		applog.Info("FindMedia: medium: ", medium)
	} else {
		applog.Error("FindMedia: ObjectId is not valid")
	}
	if titleZhCn.Valid {
		mediaMeta.TitleZhCn = titleZhCn.String
	}
	if titleEn.Valid {
		mediaMeta.TitleEn = titleEn.String
	}
	if author.Valid {
		mediaMeta.Author = author.String
	}
	if license.Valid {
		mediaMeta.License = license.String
	}
	applog.Info("FindMedia: mediaMeta ", mediaMeta)
	return &mediaMeta, nil
}
