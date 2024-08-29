package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAof(t *testing.T) {
	tempDir := t.TempDir()
	testFilePath := filepath.Join(tempDir, "test.aof")
	aof, err := NewAof(testFilePath)

	assert.Nil(t, err, "Expected no error when instantiating Aof")
	assert.NotNil(t, aof)
}

func TestNewAof_FileCreation(t *testing.T) {
	tempDir := t.TempDir()
	testFilePath := filepath.Join(tempDir, "test.aof")

	aof, err := NewAof(testFilePath)
	assert.Nil(t, err, "Expected no error creating AOF file")
	defer aof.file.Close()

	// test if file exists
	_, err = os.Stat(testFilePath)
	assert.Nil(t, err, "Expected file to be created, but it does not exist")
}

func TestNewAof_ErrorHandling(t *testing.T) {
	// Try creating file in a non-writable location
	_, err := NewAof("/root/test.aof")
	assert.Error(t, err, "Expected error when creating file fails")
}

func TestNewAof_FileSync(t *testing.T) {
	tempDir := t.TempDir()
	testFilePath := filepath.Join(tempDir, "database.aof")

	aof, err := NewAof(testFilePath)
	assert.Nil(t, err, "Expected no error creating AOF file")
	// defer aof.file.Close()

	// modify file and wait for a second to make sure it sync
	aof.mu.Lock()
	_, err = aof.file.WriteString("Test Text")
	assert.Nil(t, err, "Expected no error when writing to file")
	aof.mu.Unlock()

	time.Sleep(2 * time.Second)

	stat, err := os.Stat(testFilePath)
	assert.Nil(t, err, "Expected no error read file stats")

	assert.WithinDuration(t, time.Now(), stat.ModTime(), 3*time.Second, "Expected file to be synced with the last second")
}
