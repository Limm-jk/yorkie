/*
 * Copyright 2022 The Yorkie Authors. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package admin

import (
	"context"
	"errors"
	"fmt"
	"net"

	"google.golang.org/grpc"

	"github.com/yorkie-team/yorkie/api"
	"github.com/yorkie-team/yorkie/api/converter"
	"github.com/yorkie-team/yorkie/api/types"
	"github.com/yorkie-team/yorkie/yorkie/backend"
	"github.com/yorkie-team/yorkie/yorkie/documents"
	"github.com/yorkie-team/yorkie/yorkie/logging"
	"github.com/yorkie-team/yorkie/yorkie/projects"
)

// ErrInvalidAdminPort occurs when the port in the config is invalid.
var ErrInvalidAdminPort = errors.New("invalid port number for Admin server")

// Config is the configuration for creating a Server.
type Config struct {
	Port int `yaml:"Port"`
}

// Validate validates the port number.
func (c *Config) Validate() error {
	if c.Port < 1 || 65535 < c.Port {
		return fmt.Errorf("must be between 1 and 65535, given %d: %w", c.Port, ErrInvalidAdminPort)
	}

	return nil
}

// Server is the gRPC server for admin service.
type Server struct {
	conf       *Config
	grpcServer *grpc.Server
	backend    *backend.Backend
}

// NewServer creates a new Server.
func NewServer(conf *Config, be *backend.Backend) *Server {
	grpcServer := grpc.NewServer()

	server := &Server{
		conf:       conf,
		backend:    be,
		grpcServer: grpcServer,
	}

	api.RegisterAdminServer(grpcServer, server)

	// TODO(hackerwins): ClusterServer need to be handled by different authentication mechanism.
	// Consider extracting the servers to another grpcServer.
	api.RegisterClusterServer(grpcServer, newClusterServer(be))

	return server
}

// Start starts this server by opening the rpc port.
func (s *Server) Start() error {
	return s.listenAndServeGRPC()
}

// Shutdown shuts down this server.
func (s *Server) Shutdown(graceful bool) {
	if graceful {
		s.grpcServer.GracefulStop()
	} else {
		s.grpcServer.Stop()
	}
}

// GRPCServer returns the gRPC server.
func (s *Server) GRPCServer() *grpc.Server {
	return s.grpcServer
}

func (s *Server) listenAndServeGRPC() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.conf.Port))
	if err != nil {
		logging.DefaultLogger().Error(err)
		return err
	}

	go func() {
		logging.DefaultLogger().Infof("serving Admin on %d", s.conf.Port)

		if err := s.grpcServer.Serve(lis); err != nil {
			if err != grpc.ErrServerStopped {
				logging.DefaultLogger().Error(err)
			}
		}
	}()

	return nil
}

// CreateProject creates a new project.
func (s *Server) CreateProject(
	ctx context.Context,
	req *api.CreateProjectRequest,
) (*api.CreateProjectResponse, error) {
	project, err := projects.CreateProject(ctx, s.backend, req.Name)
	if err != nil {
		return nil, err
	}

	pbProject, err := converter.ToProject(project)
	if err != nil {
		return nil, err
	}
	return &api.CreateProjectResponse{
		Project: pbProject,
	}, nil
}

// ListProjects lists all projects.
func (s *Server) ListProjects(
	ctx context.Context,
	req *api.ListProjectsRequest,
) (*api.ListProjectsResponse, error) {
	projectList, err := projects.ListProjects(ctx, s.backend)
	if err != nil {
		return nil, err
	}

	pbProjects, err := converter.ToProjects(projectList)
	if err != nil {
		return nil, err
	}

	return &api.ListProjectsResponse{
		Projects: pbProjects,
	}, nil
}

// UpdateProject updates the project.
func (s *Server) UpdateProject(
	ctx context.Context,
	req *api.UpdateProjectRequest,
) (*api.UpdateProjectResponse, error) {
	project, err := converter.FromProject(req.Project)
	if err != nil {
		return nil, err
	}

	if err := projects.UpdateProject(
		ctx,
		s.backend,
		project,
	); err != nil {
		return nil, err
	}

	return &api.UpdateProjectResponse{}, nil
}

// ListDocuments lists documents.
func (s *Server) ListDocuments(
	ctx context.Context,
	req *api.ListDocumentsRequest,
) (*api.ListDocumentsResponse, error) {
	docs, err := documents.ListDocumentSummaries(
		ctx,
		s.backend,
		types.ID(req.PreviousId),
		int(req.PageSize),
		req.IsForward,
	)
	if err != nil {
		return nil, err
	}

	return &api.ListDocumentsResponse{
		Documents: converter.ToDocumentSummaries(docs),
	}, nil
}