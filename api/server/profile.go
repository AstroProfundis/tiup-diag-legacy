package server

import (
	"fmt"
	"os"
	"io"
	"net/http"
	"os/exec"
	"path"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pingcap/tidb-foresight/model"
	"github.com/pingcap/tidb-foresight/utils"
	log "github.com/sirupsen/logrus"
)

func (s *Server) profileAllProcess(instanceId, inspectionId string) error {
	cmd := exec.Command(
		s.config.Collector,
		fmt.Sprintf("--instance-id=%s", inspectionId),
		fmt.Sprintf("--inspection-id=%s", inspectionId),
		fmt.Sprintf("--inventory=%s", path.Join(s.config.Home, "inventory", instanceId+".ini")),
		fmt.Sprintf("--topology=%s", path.Join(s.config.Home, "topology", instanceId+".json")),
		fmt.Sprintf("--dest=%s", path.Join(s.config.Home, "inspection", inspectionId)),
		"--collect=profile",
	)
	log.Info(cmd.Args)
	err := cmd.Run()
	if err != nil {
		log.Error("run ", s.config.Collector, ": ", err)
		return err
	}
	return nil
}

func (s *Server) createProfile(r *http.Request) (*model.Inspection, error) {
	instanceId := mux.Vars(r)["id"]
	inspectionId := uuid.New().String()

	inspection := &model.Inspection{
		Uuid:       inspectionId,
		InstanceId: instanceId,
		Status:     "running",
		Type:       "manual",
	}
	err := s.model.SetInspection(inspection)
	if err != nil {
		log.Error("set inpsection: ", err)
		return nil, utils.NewForesightError(http.StatusInternalServerError, "DB_INSERT_ERROR", "error on insert data")
	}

	go func() {
		err := s.profileAllProcess(instanceId, inspectionId)
		if err != nil {
			log.Error("profile ", inspectionId, ": ", err)
			inspection.Status = "exception"
			inspection.Message = "profile failed"
			s.model.SetInspection(inspection)
			return
		}
		err = s.analyze(inspectionId)
		if err != nil {
			log.Error("analyze ", inspectionId, ": ", err)
			inspection.Status = "exception"
			inspection.Message = "analyze failed"
			s.model.SetInspection(inspection)
			return
		}
	}()

	return inspection, nil
}

func (s *Server) listProfiles(r *http.Request) (*utils.PaginationResponse, error) {
	instanceId := mux.Vars(r)["id"]
	page, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	if err != nil {
		page = 1
	}
	size, err := strconv.ParseInt(r.URL.Query().Get("per_page"), 10, 32)
	if err != nil {
		size = 10
	}
	
	profiles, total, err := s.model.ListProfiles(instanceId, page, size, path.Join(s.config.Home, "profile"))
	if err != nil {
		log.Error("list inspections: ", err)
		return nil, utils.NewForesightError(http.StatusInternalServerError, "DB_SELECT_ERROR", "error on query database")
	}

	return utils.NewPaginationResponse(total, profiles), nil
}

func (s *Server) listAllProfiles(r *http.Request) (*utils.PaginationResponse, error) {
	page, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	if err != nil {
		page = 1
	}
	size, err := strconv.ParseInt(r.URL.Query().Get("per_page"), 10, 32)
	if err != nil {
		size = 10
	}
	
	profiles, total, err := s.model.ListAllProfiles(page, size, path.Join(s.config.Home, "profile"))
	if err != nil {
		log.Error("list inspections: ", err)
		return nil, utils.NewForesightError(http.StatusInternalServerError, "DB_SELECT_ERROR", "error on query database")
	}
	return utils.NewPaginationResponse(total, profiles), nil
}

func (s *Server) getProfile(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["id"]
	comp := mux.Vars(r)["component"]
	addr := mux.Vars(r)["address"]
	tp := mux.Vars(r)["type"]
	file := mux.Vars(r)["file"]
	fpath := path.Join(s.config.Home, "profile", uuid, comp + "-" + addr, tp, file)

	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		log.Info("profile not found:", fpath)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 NOT FOUND"))
		return
	}

	localFile, err := os.Open(fpath)
	if err != nil {
		log.Error("read file: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer localFile.Close()

	io.Copy(w, localFile)
}