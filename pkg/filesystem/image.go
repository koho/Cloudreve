package filesystem

import (
	"context"
	"fmt"
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/conf"
	"github.com/HFO4/cloudreve/pkg/filesystem/fsctx"
	"github.com/HFO4/cloudreve/pkg/filesystem/response"
	"github.com/HFO4/cloudreve/pkg/thumb"
	"github.com/HFO4/cloudreve/pkg/util"
	"path/filepath"
	"strconv"
)

/* ================
     图像处理相关
   ================
*/

// HandledExtension 可以生成缩略图的文件扩展名
var HandledExtension = []string{"jpg", "jpeg", "png", "gif"}

// GetThumb 获取文件的缩略图
func (fs *FileSystem) GetThumb(ctx context.Context, id uint) (*response.ContentResponse, error) {
	// 根据 ID 查找文件
	err := fs.resetFileIDIfNotExist(ctx, id)
	if err != nil || fs.FileTarget[0].PicInfo == "" {
		return &response.ContentResponse{
			Redirect: false,
		}, ErrObjectNotExist
	}

	w, h := fs.GenerateThumbnailSize(0, 0)
	ctx = context.WithValue(ctx, fsctx.ThumbSizeCtx, [2]uint{w, h})
	ctx = context.WithValue(ctx, fsctx.FileModelCtx, fs.FileTarget[0])
	res, err := fs.Handler.Thumb(ctx, fs.GetThumbPath(&fs.FileTarget[0]))
	if err == nil && conf.SystemConfig.Mode == "master" {
		res.MaxAge = model.GetIntSetting("preview_timeout", 60)
	}

	// 出错时重新生成缩略图
	if err != nil {
		fs.GenerateThumbnail(ctx, &fs.FileTarget[0])
	}

	return res, err
}

// GenerateThumbnail 尝试为本地策略文件生成缩略图并获取图像原始大小
// TODO 失败时，如果之前还有图像信息，则清除
func (fs *FileSystem) GenerateThumbnail(ctx context.Context, file *model.File) {
	// 判断是否可以生成缩略图
	//if !IsInExtensionList(HandledExtension, file.Name) {
	//	return
	//}
	if file.Size == 0 {
		return
	}

	// 新建上下文
	newCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 获取文件数据
	source, err := fs.Handler.Get(newCtx, file.SourceName)
	if err != nil {
		return
	}
	defer source.Close()

	var image *thumb.Thumb
	var canHandle bool
	for _, thumbHandler := range thumb.Handlers {
		if thumbHandler.CanHandle(file.Name) {
			canHandle = true
			previewUrl := ""
			if thumbHandler.NeedURL() {
				if previewUrl, err = fs.GetDownloadURL(newCtx, file.ID, "doc_preview_timeout"); err != nil {
					break
				}
			}
			image, err = thumbHandler.GenerateThumb(source, util.RelativePath(file.SourceName), previewUrl)
			break
		}
	}
	if !canHandle {
		return
	}
	w, h := 1, 1
	if image != nil && err == nil {
		// 获取原始图像尺寸
		w, h = image.GetSize()

		// 生成缩略图
		image.GetThumb(fs.GenerateThumbnailSize(w, h))
		// 保存到文件
		//err = image.Save(util.RelativePath(file.SourceName + conf.ThumbConfig.FileSuffix))
		err = image.Save(util.RelativePath(fs.GetThumbPath(file)))
		if err != nil {
			util.Log().Warning("无法保存缩略图：%s", err)
			return
		}
	} else {
		util.Log().Warning("生成缩略图时无法解析 [%s] 图像数据：%s", file.SourceName, err)
	}

	// 更新文件的图像信息
	if file.Model.ID > 0 {
		err = file.UpdatePicInfo(fmt.Sprintf("%d,%d", w, h))
	} else {
		file.PicInfo = fmt.Sprintf("%d,%d", w, h)
	}

	// 失败时删除缩略图文件
	if err != nil {
		_, _ = fs.Handler.Delete(newCtx, []string{fs.GetThumbPath(file)})
	}
}

func (fs *FileSystem) GetThumbPath(file *model.File) string {
	return filepath.Join(fs.User.Policy.GenerateThumbPath(file.UserID), strconv.Itoa(int(file.UserID))+"_"+file.Name+conf.ThumbConfig.FileSuffix)
}

// GenerateThumbnailSize 获取要生成的缩略图的尺寸
func (fs *FileSystem) GenerateThumbnailSize(w, h int) (uint, uint) {
	if conf.SystemConfig.Mode == "master" {
		options := model.GetSettingByNames("thumb_width", "thumb_height")
		w, _ := strconv.ParseUint(options["thumb_width"], 10, 32)
		h, _ := strconv.ParseUint(options["thumb_height"], 10, 32)
		return uint(w), uint(h)
	}
	return conf.ThumbConfig.MaxWidth, conf.ThumbConfig.MaxHeight
}
