package types

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	"os"
	"path/filepath"
)

// Other constants
const (
	TempFileSuffix = ".temp"           // Temp file suffix
	FilePermMode   = os.FileMode(0664) // Default file permission
)

// CheckpointConfig Checkpoint configuration
type CheckpointConfig struct {
	IsEnable bool
	FilePath string
	DirPath  string
}

// SegmentPiece defines download Segment
type SegmentPiece struct {
	Index  int    // Segment number, starting from 0
	Start  int64  // Start index
	End    int64  // End index
	Offset int64  // Offset
	CRC64  uint64 // CRC check value of Segment
}

const DownloadCpMagic = "92611BED-89E2-46B6-89E5-72F273D4B0A3"

type GetObjectCheckpoint struct {
	Magic       string         // Magic
	MD5         string         // Checkpoint content MD5
	FilePath    string         // Local file
	Object      string         // Key
	ObjStat     ObjectStat     // Object status
	Segments    []SegmentPiece // All download Segments
	SegmentStat []bool         // Segments' download status
	Start       int64          // Start point of the file
	End         int64          // End point of the file
	EnableCRC   bool           // Whether has CRC check
	CRC         uint64         // CRC check value
}

// UnpackedRange
type UnpackedRange struct {
	HasStart bool  // Flag indicates if the start point is specified
	HasEnd   bool  // Flag indicates if the end point is specified
	Start    int64 // Start point
	End      int64 // End point
}

// IsValid flags of checkpoint data is valid. It returns true when the data is valid and the checkpoint is valid and the object is not updated.
func (cp GetObjectCheckpoint) IsValid(meta *storageTypes.ObjectInfo, uRange *UnpackedRange) (bool, error) {
	// Compare the CP's Magic and the MD5
	cpb := cp
	cpb.MD5 = ""
	js, _ := json.Marshal(cpb)
	sum := md5.Sum(js)
	b64 := base64.StdEncoding.EncodeToString(sum[:])

	// TODO(chris): valid cp meta
	if cp.Magic != DownloadCpMagic || b64 != cp.MD5 {
		return false, nil
	}

	// TODO(chris) LastModified Etag
	// Compare the object size, last modified time and etag
	if cp.ObjStat.Size != int64(meta.PayloadSize) {
		return false, nil
	}

	// TODO(chris): add range option

	return true, nil
}

// Load checkpoint from local file
func (cp *GetObjectCheckpoint) Load(filePath string) error {
	contents, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(contents, cp)
	return err
}

// dump funciton dumps to file
func (cp *GetObjectCheckpoint) Dump(filePath string) error {
	bcp := *cp

	// Calculate MD5
	bcp.MD5 = ""
	js, err := json.Marshal(bcp)
	if err != nil {
		return err
	}
	sum := md5.Sum(js)
	b64 := base64.StdEncoding.EncodeToString(sum[:])
	bcp.MD5 = b64

	// Serialize
	js, err = json.Marshal(bcp)
	if err != nil {
		return err
	}

	// Dump
	return os.WriteFile(filePath, js, FilePermMode)
}

// todoSegments gets unfinished Segments
func (cp GetObjectCheckpoint) TodoSegments() []SegmentPiece {
	var dps []SegmentPiece
	for i, seg := range cp.SegmentStat {
		if !seg {
			dps = append(dps, cp.Segments[i])
		}
	}
	return dps
}

// GetCompletedBytes gets completed size
func (cp GetObjectCheckpoint) GetCompletedBytes() int64 {
	var completedBytes int64
	for i, seg := range cp.Segments {
		if cp.SegmentStat[i] {
			completedBytes += (seg.End - seg.Start + 1)
		}
	}
	return completedBytes
}

// Prepare initiates download tasks
func (cp *GetObjectCheckpoint) Prepare(meta *storageTypes.ObjectInfo, objectKey, filePath string, SegmentSize int64, uRange *UnpackedRange) error {
	// CP
	cp.Magic = DownloadCpMagic
	cp.FilePath = filePath
	cp.Object = objectKey

	objectSize := meta.PayloadSize

	cp.ObjStat.Size = int64(objectSize)
	// TODO(chris) LastModified Etag

	// TODO(chris)
	//if bucket.GetConfig().IsEnableCRC && meta.Get(HTTPHeaderOssCRC64) != "" {
	//	if uRange == nil || (!uRange.HasStart && !uRange.HasEnd) {
	//		cp.enableCRC = true
	//		cp.CRC, _ = strconv.ParseUint(meta.Get(HTTPHeaderOssCRC64), 10, 64)
	//	}
	//}

	// Segments
	cp.Segments = getObjectSegments(int64(objectSize), SegmentSize, uRange)
	cp.SegmentStat = make([]bool, len(cp.Segments))
	for i := range cp.SegmentStat {
		cp.SegmentStat[i] = false
	}

	return nil
}

func (cp *GetObjectCheckpoint) Complete(cpFilePath, downFilepath string) error {
	err := os.Rename(downFilepath, cp.FilePath)
	if err != nil {
		return err
	}

	return os.Remove(cpFilePath)
}

// getDownloadsegs gets download segs
func getObjectSegments(objectSize, segSize int64, uRange *UnpackedRange) []SegmentPiece {
	segs := []SegmentPiece{}
	seg := SegmentPiece{}
	i := 0
	// TODO(chris): urange
	start := int64(0)
	end := objectSize
	//start, end := AdjustRange(uRange, objectSize)
	for offset := start; offset < end; offset += segSize {
		seg.Index = i
		seg.Start = offset
		seg.End = GetSegmentEnd(offset, end, segSize)
		seg.Offset = start
		seg.CRC64 = 0
		segs = append(segs, seg)
		i++
	}
	return segs
}

// GetSegmentEnd calculates the end position
func GetSegmentEnd(begin int64, total int64, per int64) int64 {
	if begin+per > total {
		return total - 1
	}
	return begin + per - 1
}

// AdjustRange returns adjusted range, adjust the range according to the length of the file
func AdjustRange(ur *UnpackedRange, size int64) (start, end int64) {
	if ur == nil {
		return 0, size
	}

	if ur.HasStart && ur.HasEnd {
		start = ur.Start
		end = ur.End + 1
		if ur.Start < 0 || ur.Start >= size || ur.End > size || ur.Start > ur.End {
			start = 0
			end = size
		}
	} else if ur.HasStart {
		start = ur.Start
		end = size
		if ur.Start < 0 || ur.Start >= size {
			start = 0
		}
	} else if ur.HasEnd {
		start = size - ur.End
		end = size
		if ur.End < 0 || ur.End > size {
			start = 0
			end = size
		}
	}
	return
}

func GetDownloadCpFilePath(cpConf *CheckpointConfig, srcBucket, srcObject, versionId, destFile string) string {
	if cpConf.FilePath == "" && cpConf.DirPath != "" {
		src := fmt.Sprintf("oss://%v/%v", srcBucket, srcObject)
		absPath, _ := filepath.Abs(destFile)
		cpFileName := getCpFileName(src, absPath, versionId)
		cpConf.FilePath = cpConf.DirPath + string(os.PathSeparator) + cpFileName
	}
	return cpConf.FilePath
}

// getCpFileName return the name of the checkpoint file
func getCpFileName(src, dest, versionId string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(src))
	srcCheckSum := hex.EncodeToString(md5Ctx.Sum(nil))

	md5Ctx.Reset()
	md5Ctx.Write([]byte(dest))
	destCheckSum := hex.EncodeToString(md5Ctx.Sum(nil))

	if versionId == "" {
		return fmt.Sprintf("%v-%v.cp", srcCheckSum, destCheckSum)
	}

	md5Ctx.Reset()
	md5Ctx.Write([]byte(versionId))
	versionCheckSum := hex.EncodeToString(md5Ctx.Sum(nil))
	return fmt.Sprintf("%v-%v-%v.cp", srcCheckSum, destCheckSum, versionCheckSum)
}

func CompareFiles(fileL string, fileR string) (bool, error) {
	finL, err := os.Open(fileL)
	if err != nil {
		return false, err
	}
	defer finL.Close()

	finR, err := os.Open(fileR)
	if err != nil {
		return false, err
	}
	defer finR.Close()

	statL, err := finL.Stat()
	if err != nil {
		return false, err
	}

	statR, err := finR.Stat()
	if err != nil {
		return false, err
	}

	if statL.Size() != statR.Size() {
		return false, nil
	}

	size := statL.Size()
	if size > 102400 {
		size = 102400
	}

	bufL := make([]byte, size)
	bufR := make([]byte, size)
	for {
		n, _ := finL.Read(bufL)
		if 0 == n {
			break
		}

		n, _ = finR.Read(bufR)
		if 0 == n {
			break
		}

		if !bytes.Equal(bufL, bufR) {
			return false, nil
		}
	}

	return true, nil
}
