package identify

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/stashapp/stash/pkg/models"
	"github.com/stashapp/stash/pkg/models/mocks"
	"github.com/stretchr/testify/mock"
)

func Test_createMissingStudio(t *testing.T) {
	emptyEndpoint := ""
	validEndpoint := "validEndpoint"
	invalidEndpoint := "invalidEndpoint"
	remoteSiteID := "remoteSiteID"
	validName := "validName"
	invalidName := "invalidName"
	createdID := 1

	mockStudioReaderWriter := &mocks.StudioReaderWriter{}
	mockStudioReaderWriter.On("Create", testCtx, mock.MatchedBy(func(p *models.Studio) bool {
		return p.Name == validName
	})).Run(func(args mock.Arguments) {
		s := args.Get(1).(*models.Studio)
		s.ID = createdID
	}).Return(nil)
	mockStudioReaderWriter.On("Create", testCtx, mock.MatchedBy(func(p *models.Studio) bool {
		return p.Name == invalidName
	})).Return(errors.New("error creating studio"))

	mockStudioReaderWriter.On("UpdateStashIDs", testCtx, createdID, []models.StashID{
		{
			Endpoint: invalidEndpoint,
			StashID:  remoteSiteID,
		},
	}).Return(errors.New("error updating stash ids"))
	mockStudioReaderWriter.On("UpdateStashIDs", testCtx, createdID, []models.StashID{
		{
			Endpoint: validEndpoint,
			StashID:  remoteSiteID,
		},
	}).Return(nil)

	type args struct {
		endpoint string
		studio   *models.ScrapedStudio
	}
	tests := []struct {
		name    string
		args    args
		want    *int
		wantErr bool
	}{
		{
			"simple",
			args{
				emptyEndpoint,
				&models.ScrapedStudio{
					Name: validName,
				},
			},
			&createdID,
			false,
		},
		{
			"error creating",
			args{
				emptyEndpoint,
				&models.ScrapedStudio{
					Name: invalidName,
				},
			},
			nil,
			true,
		},
		{
			"valid stash id",
			args{
				validEndpoint,
				&models.ScrapedStudio{
					Name:         validName,
					RemoteSiteID: &remoteSiteID,
				},
			},
			&createdID,
			false,
		},
		{
			"invalid stash id",
			args{
				invalidEndpoint,
				&models.ScrapedStudio{
					Name:         validName,
					RemoteSiteID: &remoteSiteID,
				},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createMissingStudio(testCtx, tt.args.endpoint, mockStudioReaderWriter, tt.args.studio)
			if (err != nil) != tt.wantErr {
				t.Errorf("createMissingStudio() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createMissingStudio() = %d, want %d", got, tt.want)
			}
		})
	}
}

func Test_scrapedToStudioInput(t *testing.T) {
	const name = "name"
	url := "url"

	tests := []struct {
		name   string
		studio *models.ScrapedStudio
		want   models.Studio
	}{
		{
			"set all",
			&models.ScrapedStudio{
				Name: name,
				URL:  &url,
			},
			models.Studio{
				Name: name,
				URL:  url,
			},
		},
		{
			"set none",
			&models.ScrapedStudio{
				Name: name,
			},
			models.Studio{
				Name: name,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scrapedToStudioInput(tt.studio)

			// clear created/updated dates
			got.CreatedAt = time.Time{}
			got.UpdatedAt = got.CreatedAt

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("scrapedToStudioInput() = %v, want %v", got, tt.want)
			}
		})
	}
}
