package archive

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	log "github.com/HiData-xyz/hit/log"
)

// Tar 目录压缩成 tar.gz 格式
// srcFile: 单个文件或者目录
func Tar(srcFile string, destTar string, trimPrefix string) error {
	var file io.Writer
	tarfile, err := os.Create(destTar)
	if err != nil {
		return err
	}
	defer tarfile.Close()
	file = tarfile

	if strings.HasSuffix(destTar, "tar.gz") {
		gw := gzip.NewWriter(tarfile)
		file = gw
		defer gw.Close()
	}

	archive := tar.NewWriter(file)
	defer archive.Close()

	filepath.Walk(srcFile, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		path = filepath.ToSlash(path)
		header.Name = strings.TrimPrefix(path, trimPrefix)
		if header.Name == "" {
			return nil
		}

		// log.Info("文件信息", "path", path, "srcFile", srcFile, "trimPrefix", trimPrefix)
		err = archive.WriteHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(archive, file)
		}
		return err
	})

	return err
}

// UnTar 解压tar文件到某个目录
// srcFile: tar压缩文件
// destDir: 解压目录
// path: 解压后得到文件目录(会以解压文件名生成新的文件夹, 放置解压文件)
func UnTar(srcFile string, destDir string) (path string, err error) {
	if !strings.HasSuffix(destDir, "/") {
		destDir = destDir + "/"
	}

	var reader io.Reader
	// 打开tar档案以供阅读。
	file, err := os.Open(srcFile)
	if err != nil {
		// tar归档结束
		log.Error("打开压缩文件失败", "err", err.Error())
		return
	}
	defer file.Close()
	reader = file
	if strings.HasSuffix(srcFile, "tar.gz") {
		gr, err := gzip.NewReader(file)
		if err != nil {
			log.Error("打开压缩文件失败", "err", err.Error())
			return "", err
		}
		reader = gr
		defer gr.Close()
	}

	tarReader := tar.NewReader(reader)
	// 迭代档案中的文件。
	for {
		tarHeader, err := tarReader.Next()
		if err != nil && err != io.EOF {
			log.Error("打开压缩文件失败", "err", err.Error())
			return "", err
		}
		if err == io.EOF {
			// tar归档结束
			break
		}

		fpath := filepath.Join(destDir, tarHeader.Name)

		if tarHeader.FileInfo().IsDir() {
			err := os.MkdirAll(fpath, os.ModePerm)
			if err != nil {
				log.Error("创建解压目录失败", "err", err.Error())
				return "", err
			}
			continue
		}

		file, err := os.Create(fpath)
		if err != nil {
			log.Error("创建文件失败", "name", fpath)
			return "", err
		}
		if _, err := io.Copy(file, tarReader); err != nil {
			log.Error("向文件写入数据失败", "name", fpath)
			return "", err
		}
		file.Close()
	}

	fileName := filepath.Base(srcFile)
	fileName = strings.TrimSuffix(fileName, ".gz") // 去除扩展名后的文件名
	fileName = strings.TrimSuffix(fileName, ".tar")
	log.Info("文件名", "srcFile", srcFile, "fileName", fileName, "destDir", destDir)
	path = fmt.Sprintf("%s%s/", destDir, fileName)
	return
}
