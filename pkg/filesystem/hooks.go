package filesystem

import (
	"context"
	"errors"
	"github.com/HFO4/cloudreve/pkg/util"
	"path"
)

// GenericBeforeUpload 通用上传前处理钩子，包含数据库操作
func GenericBeforeUpload(ctx context.Context, fs *FileSystem) error {
	file := ctx.Value(FileHeaderCtx).(FileHeader)

	// 验证单文件尺寸
	if !fs.ValidateFileSize(ctx, file.GetSize()) {
		return FileSizeTooBigError
	}

	// 验证文件名
	if !fs.ValidateLegalName(ctx, file.GetFileName()) {
		return IlegalObjectNameError
	}

	// 验证扩展名
	if !fs.ValidateExtension(ctx, file.GetFileName()) {
		return FileExtensionNotAllowedError
	}

	// 验证并扣除容量
	if !fs.ValidateCapacity(ctx, file.GetSize()) {
		return InsufficientCapacityError
	}
	return nil
}

// GenericAfterUploadCanceled 通用上传取消处理钩子，包含数据库操作
func GenericAfterUploadCanceled(ctx context.Context, fs *FileSystem) error {
	file := ctx.Value(FileHeaderCtx).(FileHeader)

	filePath := ctx.Value(SavePathCtx).(string)
	// 删除临时文件
	if util.Exists(filePath) {
		_, err := fs.Handler.Delete(ctx, []string{filePath})
		if err != nil {
			return err
		}
	}

	// 归还用户容量
	if !fs.User.DeductionStorage(file.GetSize()) {
		return errors.New("无法继续降低用户已用存储")
	}
	return nil
}

// GenericAfterUpload 文件上传完成后，包含数据库操作
func GenericAfterUpload(ctx context.Context, fs *FileSystem) error {
	// 文件存放的虚拟路径
	virtualPath := ctx.Value(FileHeaderCtx).(FileHeader).GetVirtualPath()

	// 检查路径是否存在
	isExist, folder := fs.IsPathExist(virtualPath)
	if !isExist {
		return errors.New("路径\"" + virtualPath + "\"不存在")
	}

	// 检查文件是否存在
	if fs.IsFileExist(path.Join(
		virtualPath,
		ctx.Value(FileHeaderCtx).(FileHeader).GetFileName(),
	)) {
		return errors.New("同名文件已存在")
	}

	// 向数据库中插入记录
	_, err := fs.AddFile(ctx, &folder)
	if err != nil {
		return errors.New("无法插入文件记录")
	}

	// 异步尝试生成缩略图

	return nil
}