package main

import (
	"fmt"
	"github.com/minio/minio-go/v7"
	"strconv"
	"strings"
)

func init() {
	RegisterCommand("ls", "显示所有bucket: ls、显示bucket下元素: ls [<bucket name>]", lb)
	RegisterCommand("mk", "创建Bucket: mk [<bucket name>]", createBucket)
	RegisterCommand("rm", "删除Bucket: rm [<bucket name>]、删除bucket下元素: rm [<bucket name>]/[<object name>]", rb)
	RegisterCommand("count", "计算Bucket信息: count [<bucket name>]", count)
}

func formatBytes(size int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}

	unitIndex := 0
	for size >= 1024 && unitIndex < len(units)-1 {
		size /= 1024
		unitIndex++
	}

	return strconv.FormatFloat(float64(size), 'f', 2, 64) + " " + units[unitIndex]
}

func createBucket(vars *TerminalVars, bucketName string) bool {
	if bucketName == "" {
		fmt.Println("创建失败: Bucket名称不能为空！")
	} else {
		err := client.MakeBucket(vars.Ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			fmt.Printf("创建Bucket失败: %v\n", err)
		} else {
			fmt.Println("创建Bucket成功")
		}
	}
	return true
}

func rb(vars *TerminalVars, param string) bool {
	var bucketName string
	var objectName string
	data := strings.SplitN(param, "/", 2)
	switch len(data) {
	case 1:
		bucketName = data[0]
	case 2:
		bucketName = data[0]
		objectName = data[1]
	default:
		fmt.Println("无法格式化参数")
		return true
	}
	if objectName == "" {
		return removeBucket(vars, bucketName)
	} else {
		return removeObject(vars, bucketName, objectName)
	}
}

func removeBucket(vars *TerminalVars, bucketName string) bool {
	if bucketName == "" {
		fmt.Println("删除失败：Bucket名称不能为空")
	} else {
		err := client.RemoveBucketWithOptions(vars.Ctx, bucketName, minio.RemoveBucketOptions{ForceDelete: true})
		if err != nil {
			fmt.Printf("删除Bucket失败: %v\n", err)
		} else {
			fmt.Println("删除Bucket成功")
		}
	}
	return true
}

func removeObject(vars *TerminalVars, bucketName, objectName string) bool {
	if bucketName == "" || objectName == "" {
		fmt.Println("删除Object失败：Bucket名称或Object名称不能为空")
	} else {
		err := client.RemoveObject(vars.Ctx, bucketName, objectName, minio.RemoveObjectOptions{})
		if err != nil {
			fmt.Printf("删除Object失败: %v\n", err)
		} else {
			fmt.Println("删除Object成功")
		}
	}
	return true
}

func lb(vars *TerminalVars, bucketName string) bool {
	if bucketName == "" {
		return listBuckets(vars, bucketName)
	} else {
		return listBucketObjects(vars, bucketName, true)
	}
}

func listBuckets(vars *TerminalVars, _ string) bool {
	buckets, err := client.ListBuckets(vars.Ctx)
	if err != nil {
		fmt.Println(err)
	} else {
		for _, bucket := range buckets {
			fmt.Println(bucket.Name)
		}
	}
	return true
}

func listBucketObjects(vars *TerminalVars, bucketName string, printObj bool) bool {
	if bucketName == "" {
		fmt.Println("Bucket名称不能为空")
	} else {
		var size int64
		var fileNum int64
		objects := client.ListObjects(vars.Ctx, bucketName, minio.ListObjectsOptions{
			Recursive: true,
		})
		for {
			select {
			case obj, ok := <-objects:
				if ok {
					if obj.Err == nil {
						size += obj.Size
						fileNum += 1
						if printObj {
							fmt.Println(obj.Key)
						}
					} else {
						fmt.Println(obj.Err)
						return true
					}
				} else {
					if printObj && fileNum > 0 {
						fmt.Println()
					}
					fmt.Printf("文件数量: %d, 文件大小: %s(%d)\n", fileNum, formatBytes(size), size)
					return true
				}
			}
		}
	}
	return true
}

func count(vars *TerminalVars, bucketName string) bool {
	if bucketName == "" {
		buckets, err := client.ListBuckets(vars.Ctx)
		if err != nil {
			fmt.Println(err)
		} else {
			for i, bucket := range buckets {
				if !countBucketObjects(vars, bucket.Name) {
					return false
				} else if i != len(buckets)-1 {
					fmt.Println()
				}
			}
		}
		return true
	} else {
		return countBucketObjects(vars, bucketName)
	}
}

func countBucketObjects(vars *TerminalVars, bucketName string) bool {
	if bucketName == "" {
		fmt.Println("Bucket名称不能为空")
	} else {
		fmt.Printf("Bucket: %s\n", bucketName)
		return listBucketObjects(vars, bucketName, false)
	}
	return true
}
