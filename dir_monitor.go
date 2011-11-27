//Utilities for monitoring a directory for changes
package seven5

import (
	"os"
)

//Receiver for directory change events
type DirectoryListener interface {
	FileChanged(fileInfo os.FileInfo)
	FileRemoved(fileInfo os.FileInfo)
	FileAdded(fileInfo os.FileInfo)
}

type DirectoryMonitor struct {
	Path string
	
	listeners []DirectoryListener
	previousPoll []os.FileInfo
}

func (dirMon *DirectoryMonitor) isIn(file os.FileInfo, poll []os.FileInfo) bool {
	for _, info := range poll{
		if file.Name == info.Name { return true }
	}
	return false;
}

func (dirMon *DirectoryMonitor) changed(file os.FileInfo, poll []os.FileInfo) bool {
	for _, info := range poll{
		if file.Name == info.Name { return file.Mtime_ns == info.Mtime_ns }
	}
	return true
}

func (dirMon *DirectoryMonitor) Poll() (changed bool, err error){
	dir, err := os.Open(dirMon.Path)
	if err != nil { return }

	currentPoll, err := dir.Readdir(-1)
	if err != nil { return }
	if dirMon.previousPoll == nil {
		dirMon.previousPoll = currentPoll
		return
	}
	for _, info := range currentPoll {
		if !dirMon.isIn(info, dirMon.previousPoll){
			changed = true
			for _, listener := range dirMon.listeners {
				listener.FileAdded(info)
			}
		} else if dirMon.changed(info, dirMon.previousPoll) {
			changed = true
			for _, listener := range dirMon.listeners {
				listener.FileChanged(info)
			}
		}
	}
	for _, info := range dirMon.previousPoll {
		if !dirMon.isIn(info, currentPoll){
			changed = true
			for _, listener := range dirMon.listeners {
				listener.FileRemoved(info)
			}
		}
	}
	dirMon.previousPoll = currentPoll
	return
}
func (dirMon *DirectoryMonitor) Listen(listener DirectoryListener) {
	dirMon.listeners = append(dirMon.listeners, listener)
}
func (dirMon *DirectoryMonitor) StopListening(listener *DirectoryListener) { }

func NewDirectoryMonitor(path string) (monitor *DirectoryMonitor, err error) {
	info, err := os.Stat(path)
	if err != nil { return }
	if !info.IsDirectory() { return }
	monitor = &DirectoryMonitor{Path: path}
	return
}