//go:build !windows
// +build !windows

// Copyright (c) 2015-2021 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package ioutil

import (
	"io"
	"os"
)

func isAllZeros(buffer []byte) bool {
	for _, b := range buffer {
		if b != 0 {
			return false
		}
	}
	return true
}

// AppendFile - appends the file "src" to the file "dst"
func AppendFile(dst string, src string, osync bool, off int64) error {
	flags := os.O_WRONLY | os.O_APPEND | os.O_CREATE
	if osync {
		flags |= os.O_SYNC
	}
	appendFile, err := os.OpenFile(dst, flags, 0o666)
	if err != nil {
		return err
	}
	defer appendFile.Close()

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	fileInfo, err := os.Stat(dst)
	if err != nil {
		return err
	}
	if off > 0 && off > fileInfo.Size() {
		appendFile.Truncate(off)
		appendFile.Seek(off, 0)
	}
	_, err = io.Copy(appendFile, srcFile)

	// offset := fileInfo.Size()
	// buf := make([]byte, (1 << 20))
	// for {
	// 	n, err := srcFile.Read(buf)
	// 	if err != nil && err != io.EOF {
	// 		return err
	// 	}
	// 	if n == 0 {
	// 		break
	// 	}

	// 	if !isAllZeros(buf[:n]) {
	// 		appendFile.Seek(offset, 0)
	// 		if _, err := appendFile.Write(buf[:n]); err != nil {
	// 			return err
	// 		}
	// 	} else {
	// 		// ftruncate文件
	// 		err = appendFile.Truncate(offset + int64(n))
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// 	offset += int64(n)
	// }
	return err
}
