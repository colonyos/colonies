package file

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/fs/localstore"
	"github.com/colonyos/colonies/pkg/security"
	log "github.com/sirupsen/logrus"
)

// DataServer provides access to server dependencies needed by data handlers.
type DataServer interface {
	HandleHTTPError(c backends.Context, err error, errorCode int) bool
	Validator() security.Validator
	FileDB() database.FileDatabase
	ParseSignature(payload string, signature string) (string, error)
}

// DataHandlers implements HTTP handlers for file data upload/download/delete.
type DataHandlers struct {
	server      DataServer
	objectStore localstore.ObjectStore
}

// NewDataHandlers creates a new DataHandlers instance.
func NewDataHandlers(server DataServer, objectStore localstore.ObjectStore) *DataHandlers {
	return &DataHandlers{
		server:      server,
		objectStore: objectStore,
	}
}

// authPayload is the JSON structure in the signed payload header.
type authPayload struct {
	ColonyName string `json:"colonyname"`
	ObjectName string `json:"objectname"`
	Checksum   string `json:"checksum"`
	Label      string `json:"label"`
	Name       string `json:"name"`
	Size       int64  `json:"size"`
}

// authenticate reads X-Colonies-Payload and X-Colonies-Signature headers,
// recovers the signer ID, and validates colony membership.
func (h *DataHandlers) authenticate(c backends.Context) (*authPayload, error) {
	payloadB64 := c.GetHeader("X-Colonies-Payload")
	signature := c.GetHeader("X-Colonies-Signature")

	if payloadB64 == "" || signature == "" {
		return nil, fmt.Errorf("missing authentication headers")
	}

	recoveredID, err := h.server.ParseSignature(payloadB64, signature)
	if err != nil {
		return nil, fmt.Errorf("invalid signature: %w", err)
	}

	payloadBytes, err := base64.StdEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, fmt.Errorf("invalid payload encoding: %w", err)
	}

	var payload authPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, fmt.Errorf("invalid payload JSON: %w", err)
	}

	if payload.ColonyName == "" {
		return nil, fmt.Errorf("colonyname is required in payload")
	}

	err = h.server.Validator().RequireMembership(recoveredID, payload.ColonyName, true)
	if err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	return &payload, nil
}

// HandleUpload handles PUT /api/fs/:objectName - upload a file
func (h *DataHandlers) HandleUpload(c backends.Context) {
	payload, err := h.authenticate(c)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	objectName := c.Param("objectName")
	if objectName == "" {
		h.server.HandleHTTPError(c, fmt.Errorf("objectName is required"), http.StatusBadRequest)
		return
	}

	req := c.Request()
	err = h.objectStore.Put(payload.ColonyName, objectName, req.Body, req.ContentLength)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to store file data")
		return
	}

	// Get stored file info
	info, err := h.objectStore.Stat(payload.ColonyName, objectName)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	// Register metadata in FileDB
	coloniesFile := &core.File{
		ID:          core.GenerateRandomID(),
		ColonyName:  payload.ColonyName,
		Label:       payload.Label,
		Name:        payload.Name,
		Size:        info.Size,
		Checksum:    payload.Checksum,
		ChecksumAlg: "SHA256",
		Reference: core.Reference{
			Protocol: "coloniesfs",
			S3Object: core.S3Object{Object: objectName},
		},
	}

	err = h.server.FileDB().AddFile(coloniesFile)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to register file metadata")
		return
	}

	jsonStr, err := coloniesFile.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"FileID": coloniesFile.ID, "ObjectName": objectName}).Debug("File uploaded")
	c.String(http.StatusOK, jsonStr)
}

// HandleDownload handles GET /api/fs/:objectName - download a file
func (h *DataHandlers) HandleDownload(c backends.Context) {
	payload, err := h.authenticate(c)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	objectName := c.Param("objectName")
	if objectName == "" {
		h.server.HandleHTTPError(c, fmt.Errorf("objectName is required"), http.StatusBadRequest)
		return
	}

	// Check for Range header
	rangeHeader := c.GetHeader("Range")
	w := c.ResponseWriter()

	if rangeHeader != "" {
		offset, length, err := parseRangeHeader(rangeHeader)
		if err != nil {
			h.server.HandleHTTPError(c, fmt.Errorf("invalid Range header: %w", err), http.StatusBadRequest)
			return
		}

		reader, err := h.objectStore.GetRange(payload.ColonyName, objectName, offset, length)
		if h.server.HandleHTTPError(c, err, http.StatusNotFound) {
			return
		}
		defer reader.Close()

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Accept-Ranges", "bytes")
		w.WriteHeader(http.StatusPartialContent)
		io.Copy(w, reader)
	} else {
		reader, size, err := h.objectStore.Get(payload.ColonyName, objectName)
		if h.server.HandleHTTPError(c, err, http.StatusNotFound) {
			return
		}
		defer reader.Close()

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
		w.Header().Set("Accept-Ranges", "bytes")
		w.WriteHeader(http.StatusOK)
		io.Copy(w, reader)
	}

	log.WithFields(log.Fields{"ObjectName": objectName}).Debug("File downloaded")
}

// HandleDelete handles DELETE /api/fs/:objectName - delete a file
func (h *DataHandlers) HandleDelete(c backends.Context) {
	payload, err := h.authenticate(c)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	objectName := c.Param("objectName")
	if objectName == "" {
		h.server.HandleHTTPError(c, fmt.Errorf("objectName is required"), http.StatusBadRequest)
		return
	}

	err = h.objectStore.Remove(payload.ColonyName, objectName)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to remove file data")
		return
	}

	// Also clean up any staging data
	h.objectStore.RemoveStaging(payload.ColonyName, objectName)

	log.WithFields(log.Fields{"ObjectName": objectName}).Debug("File deleted")
	c.JSON(http.StatusOK, map[string]string{"status": "deleted"})
}

// HandleChunkedInit handles POST /api/fs/:objectName/init - initiate chunked upload
func (h *DataHandlers) HandleChunkedInit(c backends.Context) {
	payload, err := h.authenticate(c)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	objectName := c.Param("objectName")
	if objectName == "" {
		h.server.HandleHTTPError(c, fmt.Errorf("objectName is required"), http.StatusBadRequest)
		return
	}

	// Clean up any existing staging data
	h.objectStore.RemoveStaging(payload.ColonyName, objectName)

	log.WithFields(log.Fields{"ObjectName": objectName}).Debug("Chunked upload initiated")
	c.JSON(http.StatusOK, map[string]string{"status": "initialized", "objectname": objectName})
}

// HandleChunkUpload handles PUT /api/fs/:objectName/:chunk - upload a single chunk
func (h *DataHandlers) HandleChunkUpload(c backends.Context) {
	payload, err := h.authenticate(c)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	objectName := c.Param("objectName")
	chunkStr := c.Param("chunk")
	if objectName == "" || chunkStr == "" {
		h.server.HandleHTTPError(c, fmt.Errorf("objectName and chunk index are required"), http.StatusBadRequest)
		return
	}

	chunkIndex, err := strconv.Atoi(chunkStr)
	if err != nil {
		h.server.HandleHTTPError(c, fmt.Errorf("invalid chunk index: %w", err), http.StatusBadRequest)
		return
	}

	req := c.Request()
	err = h.objectStore.PutChunk(payload.ColonyName, objectName, chunkIndex, req.Body, req.ContentLength)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to store chunk")
		return
	}

	log.WithFields(log.Fields{"ObjectName": objectName, "ChunkIndex": chunkIndex}).Debug("Chunk uploaded")
	c.JSON(http.StatusOK, map[string]interface{}{"status": "uploaded", "chunk": chunkIndex})
}

// HandleChunkedComplete handles POST /api/fs/:objectName/complete - finalize chunked upload
func (h *DataHandlers) HandleChunkedComplete(c backends.Context) {
	payload, err := h.authenticate(c)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	objectName := c.Param("objectName")
	if objectName == "" {
		h.server.HandleHTTPError(c, fmt.Errorf("objectName is required"), http.StatusBadRequest)
		return
	}

	// Parse total chunks from request body
	bodyBytes, err := c.ReadBody()
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	var completeReq struct {
		TotalChunks int `json:"totalchunks"`
	}
	if err := json.Unmarshal(bodyBytes, &completeReq); err != nil {
		h.server.HandleHTTPError(c, fmt.Errorf("invalid request body: %w", err), http.StatusBadRequest)
		return
	}

	if completeReq.TotalChunks <= 0 {
		h.server.HandleHTTPError(c, fmt.Errorf("totalchunks must be positive"), http.StatusBadRequest)
		return
	}

	err = h.objectStore.AssembleChunks(payload.ColonyName, objectName, completeReq.TotalChunks)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to assemble chunks")
		return
	}

	// Get stored file info and register metadata
	info, err := h.objectStore.Stat(payload.ColonyName, objectName)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	coloniesFile := &core.File{
		ID:          core.GenerateRandomID(),
		ColonyName:  payload.ColonyName,
		Label:       payload.Label,
		Name:        payload.Name,
		Size:        info.Size,
		Checksum:    payload.Checksum,
		ChecksumAlg: "SHA256",
		Reference: core.Reference{
			Protocol: "coloniesfs",
			S3Object: core.S3Object{Object: objectName},
		},
	}

	err = h.server.FileDB().AddFile(coloniesFile)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	jsonStr, err := coloniesFile.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"FileID": coloniesFile.ID, "ObjectName": objectName}).Debug("Chunked upload completed")
	c.String(http.StatusOK, jsonStr)
}

// HandleChunkStatus handles GET /api/fs/:objectName/status - check chunked upload progress
func (h *DataHandlers) HandleChunkStatus(c backends.Context) {
	payload, err := h.authenticate(c)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	objectName := c.Param("objectName")
	if objectName == "" {
		h.server.HandleHTTPError(c, fmt.Errorf("objectName is required"), http.StatusBadRequest)
		return
	}

	indices, err := h.objectStore.GetChunkStatus(payload.ColonyName, objectName)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"objectname": objectName,
		"chunks":     indices,
	})
}

// HandleSearch handles GET /api/fs - list files in a colony
func (h *DataHandlers) HandleSearch(c backends.Context) {
	payload, err := h.authenticate(c)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	objects, err := h.objectStore.List(payload.ColonyName)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"colonyname": payload.ColonyName,
		"objects":    objects,
	})
}

// parseRangeHeader parses a Range header like "bytes=0-499" and returns offset and length.
func parseRangeHeader(header string) (int64, int64, error) {
	if !strings.HasPrefix(header, "bytes=") {
		return 0, 0, fmt.Errorf("unsupported range unit")
	}

	rangeSpec := strings.TrimPrefix(header, "bytes=")
	parts := strings.SplitN(rangeSpec, "-", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid range format")
	}

	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid range start: %w", err)
	}

	if parts[1] == "" {
		// "bytes=100-" means from offset 100 to end
		return start, 0, nil
	}

	end, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid range end: %w", err)
	}

	if end < start {
		return 0, 0, fmt.Errorf("range end before start")
	}

	return start, end - start + 1, nil
}
