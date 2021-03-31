package archive_test

import (
	"errors"
	"fmt"
	"os"
	"testing"
	archive "github.com/HiData-xyz/hit/archive"
)

func TestTar(t *testing.T) {
	srcFile := "testdata"
	destTar := "testdata/small.tar"
	err := archive.Tar(srcFile, destTar, "")
	if err != nil {
		t.Error(err)
	}
	_, err = os.Stat(destTar)
	if err != nil {
		t.Error(err)
	}

	untarDir := "testdata/"
	_, err = archive.UnTar(destTar, untarDir)
	if err != nil {
		t.Error(err)
	}
	_, err = os.Stat(untarDir + srcFile + "/testdata1/small.txt")
	if err != nil {
		t.Error(err)
	}
	_, err = os.Stat(untarDir + srcFile + "/testdata1/small2.txt")
	if err != nil {
		t.Error(err)
	}
}

func TestWrapErr(t *testing.T) {
	err := errors.New("this is basic err")
	t.Log(err)
	errWrap := fmt.Errorf("this is wrap err:%w", err)
	t.Log(errWrap)

	t.Log(errors.Is(errWrap, err))

	wrap := fmt.Errorf("%w, this is wrap err", err)
	t.Log(errors.Is(wrap, err))
}
