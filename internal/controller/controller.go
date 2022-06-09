package controller

import (
	"app/internal/model"
	"app/internal/storage"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io/ioutil"
	"net/http"
)

func Build(r *chi.Mux, s *storage.Storage) {

	r.Use(middleware.Logger)

	r.Post("/users", func(w http.ResponseWriter, r *http.Request) {
		createUser(w, r, s)
	})

	r.Get("/users/{sourceID}", func(w http.ResponseWriter, r *http.Request) {
		getUser(w, r, s)
	})

	r.Patch("/users/{sourceID}", func(w http.ResponseWriter, r *http.Request) {
		updateUser(w, r, s)
	})

	r.Delete("/users/{sourceID}", func(w http.ResponseWriter, r *http.Request) {
		deleteUser(w, r, s)
	})

	r.Put("/users/{sourceID}/friends", func(w http.ResponseWriter, r *http.Request) {
		makeFriends(w, r, s)
	})

	r.Get("/users/{sourceID}/friends", func(w http.ResponseWriter, r *http.Request) {
		getFriends(w, r, s)
	})

	r.Delete("/users/{sourceID}/friends/{targetID}", func(w http.ResponseWriter, r *http.Request) {
		deleteFriend(w, r, s)
	})
}

type inputParams struct {
	TargetID string `json:"target_id"`
	Age      int    `json:"age"`
}

type message struct {
	Message string `json:"message"`
}

func readParams(w http.ResponseWriter, r *http.Request) (inputParams, error) {
	var p inputParams

	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return p, err
	}
	defer r.Body.Close()

	if err := json.Unmarshal(content, &p); err != nil {
		return p, err
	}

	return p, nil
}
func createUser(w http.ResponseWriter, r *http.Request, s *storage.Storage) {
	w.Header().Add("Content-Type", "application/json")
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		b, _ := json.Marshal(message{err.Error()})
		w.Write(b)
		return
	}
	defer r.Body.Close()

	var u model.User

	if err := json.Unmarshal(content, &u); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		b, _ := json.Marshal(message{err.Error()})
		w.Write(b)
		return
	}

	userID, err := s.PutUser(u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		b, _ := json.Marshal(message{err.Error()})
		w.Write(b)
		return
	}

	result := message{"Username_" + userID + " was created"}
	b, _ := json.Marshal(result)
	w.Write(b)
	w.WriteHeader(http.StatusCreated)
}

func getUser(w http.ResponseWriter, r *http.Request, s *storage.Storage) {
	w.Header().Add("Content-Type", "application/json")

	sourceID := chi.URLParam(r, "sourceID")

	u, err := s.GetUser(sourceID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		b, _ := json.Marshal(message{err.Error()})
		w.Write(b)
		return
	}

	b, _ := json.Marshal(u)
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func makeFriends(w http.ResponseWriter, r *http.Request, s *storage.Storage) {
	w.Header().Add("Content-Type", "application/json")
	sourceID := chi.URLParam(r, "sourceID")

	p, err := readParams(w, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		b, _ := json.Marshal(message{err.Error()})
		w.Write(b)
		return
	}

	if err := s.MakeFriends(sourceID, p.TargetID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		b, _ := json.Marshal(message{err.Error()})
		w.Write(b)
		return
	}

	result := message{"Username_" + p.TargetID + " и username_" + sourceID + " теперь друзья"}
	b, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func deleteUser(w http.ResponseWriter, r *http.Request, s *storage.Storage) {
	w.Header().Add("Content-Type", "application/json")
	sourceID := chi.URLParam(r, "sourceID")

	if err := s.DeleteUser(sourceID); err != nil {
		w.WriteHeader(http.StatusNotFound)
		b, _ := json.Marshal(message{err.Error()})
		w.Write(b)
		return
	}

	result := message{"Username_" + sourceID + " удален"}
	b, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func getFriends(w http.ResponseWriter, r *http.Request, s *storage.Storage) {
	w.Header().Add("Content-Type", "application/json")

	sourceID := chi.URLParam(r, "sourceID")

	friends, err := s.GetFriends(sourceID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		b, _ := json.Marshal(message{err.Error()})
		w.Write(b)
		return
	}

	b, _ := json.Marshal(friends)
	b, _ = json.Marshal(message{string(b)})
	w.Write(b)
}

func updateUser(w http.ResponseWriter, r *http.Request, s *storage.Storage) {
	w.Header().Add("Content-Type", "application/json")
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		b, _ := json.Marshal(message{err.Error()})
		w.Write(b)
		return
	}
	defer r.Body.Close()

	var u model.User

	if err := json.Unmarshal(content, &u); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		b, _ := json.Marshal(message{err.Error()})
		w.Write(b)
		return
	}

	sourceID := chi.URLParam(r, "sourceID")

	if err := s.Update(sourceID, u); err != nil {
		w.WriteHeader(http.StatusNotFound)
		b, _ := json.Marshal(message{err.Error()})
		w.Write(b)
		return
	}

	result := message{"Данные пользователя Username_" + sourceID + " успешно обновлены"}
	b, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func deleteFriend(w http.ResponseWriter, r *http.Request, s *storage.Storage) {
	w.Header().Add("Content-Type", "application/json")
	sourceID := chi.URLParam(r, "sourceID")
	targetID := chi.URLParam(r, "targetID")

	if err := s.DeleteFriend(sourceID, targetID); err != nil {
		w.WriteHeader(http.StatusNotFound)
		b, _ := json.Marshal(message{err.Error()})
		w.Write(b)
		return
	}

	result := message{"Username_" + targetID + " и username_" + sourceID + " больше не друзья"}
	b, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
