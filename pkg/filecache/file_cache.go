package filecache

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/yunhanshu-net/runcher/pkg/osx"
	"sync"
	"time"
)

// FileCache 初期采用本地缓存，后期考虑换成minio或者其他分布式存储
type FileCache interface {
	Get(ossPath string, addExpire time.Duration) (file *File, exist bool)
	Set(ossPath string, localPath string, ttl time.Duration) (cover bool)
	Del(ossPath string) bool
	DeleteTask(path string, ttl time.Duration) //定时删除文件
	Close()
}

type File struct {
	filePath   string //存储的目标地址
	expireTime int64  //到期需要移除的时间

}

type deleteTask struct {
	//mutex      *sync.Mutex
	expireTime int //时间戳
}

type LocalFileCache struct {
	mutex      *sync.Mutex
	closeChan  chan struct{}
	fileMap    map[string]*File
	deleteTask map[string]*deleteTask //时间戳/ 文件列表 过期的文件需要删除
	checkLimit time.Duration
	tk         *time.Ticker
}

func (c *LocalFileCache) check() {
	//巡检任务
	for {
		select {
		case <-c.closeChan:
			fmt.Println("local file file cache close")
			return
		case <-c.tk.C:
			for s, file := range c.fileMap {
				if file.expireTime == -1 {
					continue
				}
				if time.Now().Unix() > file.expireTime { //说明已经过期需要删除文件
					c.mutex.Lock()
					p := file.filePath
					logrus.Infof("删除文件：%s\n", p)
					go osx.DeleteFileOrDir(p)
					delete(c.fileMap, s)
					c.mutex.Unlock()
				}
			}
			for file, task := range c.deleteTask { //移除到期文件
				if time.Now().Unix() > int64(task.expireTime) {
					logrus.Infof("删除文件：%s\n", file)
					go osx.DeleteFileOrDir(file)
					delete(c.deleteTask, file)
				}
			}
		default:

		}

	}
}

func NewLocalFileCache() *LocalFileCache {
	l := &LocalFileCache{
		closeChan:  make(chan struct{}, 1),
		mutex:      &sync.Mutex{},
		checkLimit: time.Second * 5,
		fileMap:    make(map[string]*File),
		tk:         time.NewTicker(time.Second * 5),
		deleteTask: map[string]*deleteTask{},
	}
	go l.check()
	return l
}

func (c *LocalFileCache) Get(path string, addExpire time.Duration) (file *File, exist bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	f, ok := c.fileMap[path]
	if ok {
		if addExpire.Microseconds() > 0 {
			f.expireTime = time.Now().Add(addExpire).Unix()
		}
	} else {
		return nil, false
	}

	return f, ok
}

func (c *LocalFileCache) Set(path string, distPath string, ttl time.Duration) (cover bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	f := File{
		filePath:   distPath,
		expireTime: time.Now().Add(ttl).Unix(),
	}
	_, ok := c.fileMap[path]
	c.fileMap[path] = &f
	return ok
}

func (c *LocalFileCache) Del(path string) (exist bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	_, ok := c.fileMap[path]
	if ok {
		delete(c.fileMap, path)
		return true
	}
	return false
}
func (c *LocalFileCache) Close() {
	c.closeChan <- struct{}{}
}

// DeleteTask 删除任务
func (c *LocalFileCache) DeleteTask(path string, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	ts := time.Now().Add(ttl).Unix()
	c.deleteTask[path] = &deleteTask{
		//mutex:      &sync.Mutex{},
		expireTime: int(ts),
	}
}
