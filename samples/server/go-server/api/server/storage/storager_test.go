// Copyright 2017 The Grafeas Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package storage

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/grafeas/grafeas/samples/server/go-server/api/server/name"
	"github.com/grafeas/grafeas/samples/server/go-server/api/server/testing"
	server "github.com/grafeas/grafeas/server-go"

	"reflect"
	"strings"
	"testing"

	pb "github.com/grafeas/grafeas/v1alpha1/proto"
	opspb "google.golang.org/genproto/googleapis/longrunning"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Tests implementations of server.Storager
// createStore is a function that creates new server.Storage instances and
// a corresponding cleanUp function that will be run at the end of each
// test case.
func doTestStorager(t *testing.T, createStore func(t *testing.T) (server.Storager, func())) {
	t.Run("CreateProject", func(t *testing.T) {
		s, cleanUp := createStore(t)
		defer cleanUp()
		p := "myproject"
		if err := s.CreateProject(p); err != nil {
			t.Errorf("CreateProject got %v want success", err)
		}
		// Try to insert the same project twice, expect failure.
		if err := s.CreateProject(p); err == nil {
			t.Errorf("CreateProject got success, want Error")
		} else if s, _ := status.FromError(err); s.Code() != codes.AlreadyExists {
			t.Errorf("CreateProject got code %v want %v", s.Code(), codes.AlreadyExists)
		}
	})

	t.Run("CreateNote", func(t *testing.T) {
		s, cleanUp := createStore(t)
		defer cleanUp()
		nPID := "vulnerability-scanner-a"
		n := testutil.Note(nPID)
		if err := s.CreateNote(n); err != nil {
			t.Errorf("CreateNote got %v want success", err)
		}
		// Try to insert the same note twice, expect failure.
		if err := s.CreateNote(n); err == nil {
			t.Errorf("CreateNote got success, want Error")
		} else if s, _ := status.FromError(err); s.Code() != codes.AlreadyExists {
			t.Errorf("CreateNote got code %v want %v", s.Code(), codes.AlreadyExists)
		}
	})

	t.Run("CreateOccurrence", func(t *testing.T) {
		s, cleanUp := createStore(t)
		defer cleanUp()
		nPID := "vulnerability-scanner-a"
		n := testutil.Note(nPID)
		if err := s.CreateNote(n); err != nil {
			t.Fatalf("CreateNote got %v want success", err)
		}
		oPID := "occurrence-project"
		o := testutil.Occurrence(oPID, n.Name)
		if err := s.CreateOccurrence(o); err != nil {
			t.Errorf("CreateOccurrence got %v want success", err)
		}
		// Try to insert the same occurrence twice, expect failure.
		if err := s.CreateOccurrence(o); err == nil {
			t.Errorf("CreateOccurrence got success, want Error")
		} else if s, _ := status.FromError(err); s.Code() != codes.AlreadyExists {
			t.Errorf("CreateOccurrence got code %v want %v", s.Code(), codes.AlreadyExists)
		}
		pID, oID, err := name.ParseOccurrence(o.Name)
		if err != nil {
			t.Fatalf("Error parsing projectID and occurrenceID %v", err)
		}
		if got, err := s.GetOccurrence(pID, oID); err != nil {
			t.Fatalf("GetOccurrence got %v, want success", err)
		} else if !reflect.DeepEqual(got, o) {
			t.Errorf("GetOccurrence got %v, want %v", got, o)
		}
	})

	t.Run("CreateOperation", func(t *testing.T) {
		s, cleanUp := createStore(t)
		defer cleanUp()
		opPID := "vulnerability-scanner-a"
		op := testutil.Operation(opPID)
		if err := s.CreateOperation(op); err != nil {
			t.Errorf("CreateOperation got %v want success", err)
		}
		// Try to insert the same note twice, expect failure.
		if err := s.CreateOperation(op); err == nil {
			t.Errorf("CreateOperation got success, want Error")
		} else if s, _ := status.FromError(err); s.Code() != codes.AlreadyExists {
			t.Errorf("CreateOperation got code %v want %v", s.Code(), codes.AlreadyExists)
		}
	})

	t.Run("DeleteProject", func(t *testing.T) {
		s, cleanUp := createStore(t)
		defer cleanUp()
		pID := "myproject"
		// Delete before the note exists
		if err := s.DeleteProject(pID); err == nil {
			t.Error("Deleting nonexistant note got success, want error")
		}
		if err := s.CreateProject(pID); err != nil {
			t.Fatalf("CreateProject got %v want success", err)
		}

		if err := s.DeleteProject(pID); err != nil {
			t.Errorf("DeleteProject got %v, want success ", err)
		}
	})

	t.Run("DeleteOccurrence", func(t *testing.T) {
		s, cleanUp := createStore(t)
		defer cleanUp()
		nPID := "vulnerability-scanner-a"
		n := testutil.Note(nPID)
		if err := s.CreateNote(n); err != nil {
			t.Fatalf("CreateNote got %v want success", err)
		}
		oPID := "occurrence-project"
		o := testutil.Occurrence(oPID, n.Name)
		// Delete before the occurrence exists
		pID, oID, err := name.ParseOccurrence(o.Name)
		if err != nil {
			t.Fatalf("Error parsing occurrence %v", err)
		}
		if err := s.DeleteOccurrence(pID, oID); err == nil {
			t.Error("Deleting nonexistant occurrence got success, want error")
		}
		if err := s.CreateOccurrence(o); err != nil {
			t.Fatalf("CreateOccurrence got %v want success", err)
		}
		if err := s.DeleteOccurrence(pID, oID); err != nil {
			t.Errorf("DeleteOccurrence got %v, want success ", err)
		}
	})

	t.Run("UpdateOccurrence", func(t *testing.T) {
		s, cleanUp := createStore(t)
		defer cleanUp()
		nPID := "vulnerability-scanner-a"
		n := testutil.Note(nPID)
		if err := s.CreateNote(n); err != nil {
			t.Fatalf("CreateNote got %v want success", err)
		}
		oPID := "occurrence-project"
		o := testutil.Occurrence(oPID, n.Name)
		pID, oID, err := name.ParseOccurrence(o.Name)
		if err != nil {
			t.Fatalf("Error parsing projectID and occurrenceID %v", err)
		}
		if err := s.UpdateOccurrence(pID, oID, o); err == nil {
			t.Fatal("UpdateOccurrence got success want error")
		}
		if err := s.CreateOccurrence(o); err != nil {
			t.Fatalf("CreateOccurrence got %v want success", err)
		}
		if got, err := s.GetOccurrence(pID, oID); err != nil {
			t.Fatalf("GetOccurrence got %v, want success", err)
		} else if !reflect.DeepEqual(got, o) {
			t.Errorf("GetOccurrence got %v, want %v", got, o)
		}

		o2 := o
		o2.GetVulnerabilityDetails().CvssScore = 1.0
		if err := s.UpdateOccurrence(pID, oID, o2); err != nil {
			t.Fatalf("UpdateOccurrence got %v want success", err)
		}

		if got, err := s.GetOccurrence(pID, oID); err != nil {
			t.Fatalf("GetOccurrence got %v, want success", err)
		} else if !reflect.DeepEqual(got, o2) {
			t.Errorf("GetOccurrence got %v, want %v", got, o2)
		}
	})

	t.Run("DeleteNote", func(t *testing.T) {
		s, cleanUp := createStore(t)
		defer cleanUp()
		nPID := "vulnerability-scanner-a"
		n := testutil.Note(nPID)
		// Delete before the note exists
		pID, oID, err := name.ParseNote(n.Name)
		if err != nil {
			t.Fatalf("Error parsing note %v", err)
		}
		if err := s.DeleteNote(pID, oID); err == nil {
			t.Error("Deleting nonexistant note got success, want error")
		}
		if err := s.CreateNote(n); err != nil {
			t.Fatalf("CreateNote got %v want success", err)
		}

		if err := s.DeleteNote(pID, oID); err != nil {
			t.Errorf("DeleteNote got %v, want success ", err)
		}
	})

	t.Run("UpdateNote", func(t *testing.T) {
		s, cleanUp := createStore(t)
		defer cleanUp()
		nPID := "vulnerability-scanner-a"
		n := testutil.Note(nPID)

		pID, nID, err := name.ParseNote(n.Name)
		if err != nil {
			t.Fatalf("Error parsing projectID and noteID %v", err)
		}
		if err := s.UpdateNote(pID, nID, n); err == nil {
			t.Fatal("UpdateNote got success want error")
		}
		if err := s.CreateNote(n); err != nil {
			t.Fatalf("CreateNote got %v want success", err)
		}
		if got, err := s.GetNote(pID, nID); err != nil {
			t.Fatalf("GetNote got %v, want success", err)
		} else if !reflect.DeepEqual(got, n) {
			t.Errorf("GetNote got %v, want %v", got, n)
		}

		n2 := n
		n2.GetVulnerabilityType().CvssScore = 1.0
		if err := s.UpdateNote(pID, nID, n2); err != nil {
			t.Fatalf("UpdateNote got %v want success", err)
		}

		if got, err := s.GetNote(pID, nID); err != nil {
			t.Fatalf("GetNote got %v, want success", err)
		} else if !reflect.DeepEqual(got, n2) {
			t.Errorf("GetNote got %v, want %v", got, n2)
		}
	})

	t.Run("GetProject", func(t *testing.T) {
		s, cleanUp := createStore(t)
		defer cleanUp()
		pID := "myproject"
		// Try to get project before it has been created, expect failure.
		if _, err := s.GetProject(pID); err == nil {
			t.Errorf("GetProject got success, want Error")
		} else if s, _ := status.FromError(err); s.Code() != codes.NotFound {
			t.Errorf("GetProject got code %v want %v", s.Code(), codes.NotFound)
		}
		s.CreateProject(pID)
		if p, err := s.GetProject(pID); err != nil {
			t.Fatalf("GetProject got %v want success", err)
		} else if p.Name != name.FormatProject(pID) {
			t.Fatalf("Got %s want %s", p.Name, pID)
		}
	})

	t.Run("GetOccurrence", func(t *testing.T) {
		s, cleanUp := createStore(t)
		defer cleanUp()
		nPID := "vulnerability-scanner-a"
		n := testutil.Note(nPID)
		if err := s.CreateNote(n); err != nil {
			t.Fatalf("CreateNote got %v want success", err)
		}
		oPID := "occurrence-project"
		o := testutil.Occurrence(oPID, n.Name)
		pID, oID, err := name.ParseOccurrence(o.Name)
		if err != nil {
			t.Fatalf("Error parsing occurrence %v", err)
		}
		if _, err := s.GetOccurrence(pID, oID); err == nil {
			t.Fatal("GetOccurrence got success, want error")
		}
		if err := s.CreateOccurrence(o); err != nil {
			t.Errorf("CreateOccurrence got %v, want Success", err)
		}
		if got, err := s.GetOccurrence(pID, oID); err != nil {
			t.Fatalf("GetOccurrence got %v, want success", err)
		} else if !reflect.DeepEqual(got, o) {
			t.Errorf("GetOccurrence got %v, want %v", got, o)
		}
	})

	t.Run("GetNote", func(t *testing.T) {
		s, cleanUp := createStore(t)
		defer cleanUp()
		nPID := "vulnerability-scanner-a"
		n := testutil.Note(nPID)

		pID, nID, err := name.ParseNote(n.Name)
		if err != nil {
			t.Fatalf("Error parsing note %v", err)
		}
		if _, err := s.GetNote(pID, nID); err == nil {
			t.Fatal("GetNote got success, want error")
		}
		if err := s.CreateNote(n); err != nil {
			t.Errorf("CreateNote got %v, want Success", err)
		}
		if got, err := s.GetNote(pID, nID); err != nil {
			t.Fatalf("GetNote got %v, want success", err)
		} else if !reflect.DeepEqual(got, n) {
			t.Errorf("GetNote got %v, want %v", got, n)
		}
	})

	t.Run("GetNoteByOccurrence", func(t *testing.T) {
		s, cleanUp := createStore(t)
		defer cleanUp()
		nPID := "vulnerability-scanner-a"
		n := testutil.Note(nPID)
		if err := s.CreateNote(n); err != nil {
			t.Fatalf("CreateNote got %v want success", err)
		}
		oPID := "occurrence-project"
		o := testutil.Occurrence(oPID, n.Name)
		pID, oID, err := name.ParseOccurrence(o.Name)
		if err != nil {
			t.Fatalf("Error parsing occurrence %v", err)
		}
		if _, err := s.GetNoteByOccurrence(pID, oID); err == nil {
			t.Fatal("GetNoteByOccurrence got success, want error")
		}
		if err := s.CreateOccurrence(o); err != nil {
			t.Errorf("CreateOccurrence got %v, want Success", err)
		}
		if got, err := s.GetNoteByOccurrence(pID, oID); err != nil {
			t.Fatalf("GetNoteByOccurrence got %v, want success", err)
		} else if !reflect.DeepEqual(got, n) {
			t.Errorf("GetNoteByOccurrence got %v, want %v", got, n)
		}
	})

	t.Run("GetOperation", func(t *testing.T) {
		s, cleanUp := createStore(t)
		defer cleanUp()
		oPID := "vulnerability-scanner-a"
		o := testutil.Operation(oPID)

		pID, oID, err := name.ParseOperation(o.Name)
		if err != nil {
			t.Fatalf("Error parsing operation %v", err)
		}
		if _, err := s.GetOperation(pID, oID); err == nil {
			t.Fatal("GetOperation got success, want error")
		}
		if err := s.CreateOperation(o); err != nil {
			t.Errorf("CreateOperation got %v, want Success", err)
		}
		if got, err := s.GetOperation(pID, oID); err != nil {
			t.Fatalf("GetOperation got %v, want success", err)
		} else if !reflect.DeepEqual(got, o) {
			t.Errorf("GetOperation got %v, want %v", got, o)
		}
	})

	t.Run("DeleteOperation", func(t *testing.T) {
		s, cleanUp := createStore(t)
		defer cleanUp()
		oPID := "vulnerability-scanner-a"
		o := testutil.Operation(oPID)
		// Delete before the operation exists
		pID, oID, err := name.ParseOperation(o.Name)
		if err != nil {
			t.Fatalf("Error parsing note %v", err)
		}
		if err := s.DeleteOperation(pID, oID); err == nil {
			t.Error("Deleting nonexistant operation got success, want error")
		}
		if err := s.CreateOperation(o); err != nil {
			t.Fatalf("CreateOperation got %v want success", err)
		}

		if err := s.DeleteOperation(pID, oID); err != nil {
			t.Errorf("DeleteOperation got %v, want success ", err)
		}
	})

	t.Run("UpdateOperation", func(t *testing.T) {
		s, cleanUp := createStore(t)
		defer cleanUp()
		oPID := "vulnerability-scanner-a"
		o := testutil.Operation(oPID)

		pID, oID, err := name.ParseOperation(o.Name)
		if err != nil {
			t.Fatalf("Error parsing projectID and operationID %v", err)
		}
		if err := s.UpdateOperation(pID, oID, o); err == nil {
			t.Fatal("UpdateOperation got success want error")
		}
		if err := s.CreateOperation(o); err != nil {
			t.Fatalf("CreateOperation got %v want success", err)
		}
		if got, err := s.GetOperation(pID, oID); err != nil {
			t.Fatalf("GetOperation got %v, want success", err)
		} else if !reflect.DeepEqual(got, o) {
			t.Errorf("GetOperation got %v, want %v", got, o)
		}

		o2 := o
		o2.Done = true
		if err := s.UpdateOperation(pID, oID, o2); err != nil {
			t.Fatalf("UpdateOperation got %v want success", err)
		}

		if got, err := s.GetOperation(pID, oID); err != nil {
			t.Fatalf("GetOperation got %v, want success", err)
		} else if !reflect.DeepEqual(got, o2) {
			t.Errorf("GetOperation got %v, want %v", got, o2)
		}
	})

	t.Run("ListProjects", func(t *testing.T) {
		s, cleanUp := createStore(t)
		defer cleanUp()
		wantProjectNames := []string{}
		for i := 0; i < 20; i++ {
			pID := fmt.Sprint("Project", i)
			if err := s.CreateProject(pID); err != nil {
				t.Fatalf("CreateProject got %v want success", err)
			}
			wantProjectNames = append(wantProjectNames, name.FormatProject(pID))
		}
		filter := "filters_are_yet_to_be_implemented"
		gotProjects, err := s.ListProjects(filter)
		if err != nil {
			t.Fatalf("ListProjects got %v want success", err)
		}
		if len(gotProjects) != 20 {
			t.Errorf("ListProjects got %v operations, want 20", len(gotProjects))
		}
		gotProjectNames := make([]string, len(gotProjects))
		for i, project := range gotProjects {
			gotProjectNames[i] = project.Name
		}
		// Sort to handle that wantProjectNames are not guaranteed to be listed in insertion order
		sort.Strings(wantProjectNames)
		sort.Strings(gotProjectNames)
		if !reflect.DeepEqual(gotProjectNames, wantProjectNames) {
			t.Errorf("ListProjects got %v want %v", gotProjectNames, wantProjectNames)
		}
	})

	t.Run("ListOperations", func(t *testing.T) {
		s, cleanUp := createStore(t)
		defer cleanUp()
		ops := []opspb.Operation{}
		findProject := "findThese"
		dontFind := "dontFind"
		for i := 0; i < 20; i++ {
			o := testutil.Operation("")
			if i < 5 {
				o.Name = name.FormatOperation(findProject, strconv.Itoa(i))
			} else {
				o.Name = name.FormatOperation(dontFind, strconv.Itoa(i))
			}
			if err := s.CreateOperation(o); err != nil {
				t.Fatalf("CreateOperation got %v want success", err)
			}
			ops = append(ops, *o)
		}
		gotOs, err := s.ListOperations(findProject, "")
		if err != nil {
			t.Fatalf("ListOperations got %v want success", err)
		}

		if len(gotOs) != 5 {
			t.Errorf("ListOperations got %v operations, want 5", len(gotOs))
		}
		for _, o := range gotOs {
			want := name.FormatProject(findProject)
			if !strings.HasPrefix(o.Name, want) {
				t.Errorf("ListOperations got %v want prefix %v", o.Name, want)
			}
		}
	})

	t.Run("ListNotes", func(t *testing.T) {
		s, cleanUp := createStore(t)
		defer cleanUp()
		ns := []*pb.Note{}
		findProject := "findThese"
		dontFind := "dontFind"
		for i := 0; i < 20; i++ {
			n := testutil.Note("")
			if i < 5 {
				n.Name = name.FormatNote(findProject, strconv.Itoa(i))
			} else {
				n.Name = name.FormatNote(dontFind, strconv.Itoa(i))
			}
			if err := s.CreateNote(n); err != nil {
				t.Fatalf("CreateNote got %v want success", err)
			}
			ns = append(ns, n)
		}
		gotNs, err := s.ListNotes(findProject, "")
		if err != nil {
			t.Fatalf("ListNotes got %v want success", err)
		}
		if len(gotNs) != 5 {
			t.Errorf("ListNotes got %v operations, want 5", len(gotNs))
		}
		for _, n := range gotNs {
			want := name.FormatProject(findProject)
			if !strings.HasPrefix(n.Name, want) {
				t.Errorf("ListNotes got %v want %v", n.Name, want)
			}
		}
	})

	t.Run("ListOccurrences", func(t *testing.T) {
		s, cleanUp := createStore(t)
		defer cleanUp()
		os := []*pb.Occurrence{}
		findProject := "findThese"
		dontFind := "dontFind"
		nPID := "vulnerability-scanner-a"
		n := testutil.Note(nPID)
		if err := s.CreateNote(n); err != nil {
			t.Fatalf("CreateNote got %v want success", err)
		}
		for i := 0; i < 20; i++ {
			oPID := "_"
			o := testutil.Occurrence(oPID, n.Name)
			if i < 5 {
				o.Name = name.FormatOccurrence(findProject, strconv.Itoa(i))
			} else {
				o.Name = name.FormatOccurrence(dontFind, strconv.Itoa(i))
			}
			if err := s.CreateOccurrence(o); err != nil {
				t.Fatalf("CreateOccurrence got %v want success", err)
			}
			os = append(os, o)
		}
		gotOs, err := s.ListOccurrences(findProject, "")
		if err != nil {
			t.Fatalf("ListOccurrences got %v want success", err)
		}
		if len(gotOs) != 5 {
			t.Errorf("ListOccurrences got %v Occurrences, want 5", len(gotOs))
		}
		for _, o := range gotOs {
			want := name.FormatProject(findProject)
			if !strings.HasPrefix(o.Name, want) {
				t.Errorf("ListOccurrences got %v want  %v", o.Name, want)
			}
		}
	})

	t.Run("ListNoteOccurrences", func(t *testing.T) {
		s, cleanUp := createStore(t)
		defer cleanUp()
		os := []*pb.Occurrence{}
		findProject := "findThese"
		dontFind := "dontFind"
		nPID := "vulnerability-scanner-a"
		n := testutil.Note(nPID)
		if err := s.CreateNote(n); err != nil {
			t.Fatalf("CreateNote got %v want success", err)
		}
		for i := 0; i < 20; i++ {
			oPID := "_"
			o := testutil.Occurrence(oPID, n.Name)
			if i < 5 {
				o.Name = name.FormatOccurrence(findProject, strconv.Itoa(i))
			} else {
				o.Name = name.FormatOccurrence(dontFind, strconv.Itoa(i))
			}
			if err := s.CreateOccurrence(o); err != nil {
				t.Fatalf("CreateOccurrence got %v want success", err)
			}
			os = append(os, o)
		}
		pID, nID, err := name.ParseNote(n.Name)
		if err != nil {
			t.Fatalf("Error parsing note name %v", err)
		}
		gotOs, err := s.ListNoteOccurrences(pID, nID, "")
		if err != nil {
			t.Fatalf("ListNoteOccurrences got %v want success", err)
		}
		if len(gotOs) != 20 {
			t.Errorf("ListNoteOccurrences got %v Occurrences, want 20", len(gotOs))
		}
		for _, o := range gotOs {
			if o.NoteName != n.Name {
				t.Errorf("ListNoteOccurrences got %v want  %v", o.Name, o.NoteName)
			}
		}
	})
}
