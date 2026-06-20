package converter

import (
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/postgres/model"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/entity"
)

//go:generate goverter gen .

// goverter:converter
// goverter:output:file avatar_gen.go
// goverter:output:package converter
// goverter:extend github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/postgres/converter/convext:CopyTime
// goverter:extend github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/postgres/converter/convext:CopyTimePtr
// goverter:extend github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/postgres/converter/convext:UUIDToString
// goverter:extend github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/postgres/converter/convext:StringToUUID
// goverter:extend github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/postgres/converter/convext:StringToUploadStatus
// goverter:extend github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/postgres/converter/convext:UploadStatusToString
// goverter:extend github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/postgres/converter/convext:StringToProcessingStatus
// goverter:extend github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/postgres/converter/convext:ProcessingStatusToString
// goverter:extend github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/postgres/converter/convext:RawMessageToThumbnailS3Keys
// goverter:extend github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/postgres/converter/convext:ThumbnailS3KeysToRawMessage
type AvatarConverter interface {
	ToEntity(source model.Avatar) (entity.Avatar, error)
	ToModel(source entity.Avatar) (model.Avatar, error)
}
