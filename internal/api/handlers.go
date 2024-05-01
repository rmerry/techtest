package api

import (
	"atostechtest/internal/encryption"
	"atostechtest/internal/sessionstore"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Handlers struct {
	sessionStore *sessionstore.Store
	Router       chi.Router
	logger       *slog.Logger
}

// @title			Richard Merry ATOS Tech Test
// @description	A simple API for creating symmetric encryption sessions within which plaintext can be encrypted and cipher text decrypted. Sessions have a limited lifetime, currently set to 10 minutes.
// @contact.name	Richard Merry
// @host			localhost:8081
// @BasePath		/api/v1
func NewHTTPHandlers(sessionStore *sessionstore.Store) *Handlers {
	logger := slog.Default().With("component", "api")
	mux := chi.NewRouter()

	h := &Handlers{
		sessionStore: sessionStore,
		Router:       mux,
		logger:       logger,
	}

	// Attach middleware.
	mux.Use(middleware.RequestID)
	mux.Use(NewSlogRequestLogger(logger.Handler()))

	mux.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {

			r.Route("/session", func(r chi.Router) {
				r.Post("/", h.createSession)

				r.Route("/{sessionID}", func(r chi.Router) {
					r.Use(h.sessionCtx) // Put the session on the request context.

					r.Route("/encrypt", func(r chi.Router) {
						r.Post("/", h.createEncrypt)
					})
					r.Route("/decrypt", func(r chi.Router) {
						r.Post("/", h.createDecrypt)
					})
				})
			})

			r.Route("/algorithms", func(r chi.Router) {
				r.Get("/", h.getAlgorithms)
			})

		})
	})

	return h
}

// Decrypts a base64 encoded cipher text input and return an unencoded plaintext.
//
//	@Summary		Decrypt cipher text.
//	@Description	Decrypt cipher text in the context of a specific encryption session.
//	@Description	The cipher will be decrypted using the specific algorithm and key associated with the session.
//	@Tags			encryption, session
//	@Accept			json
//	@Produce		json
//	@Param			session_id	path		string			false	"An encryption session ID"
//	@Param			request		body		DecryptRequest	true	"Request body"
//	@Success		200			{object}	DecryptResponse
//	@Failure		400			{object}	ErrResponse
//	@Failure		404			{object}	ErrResponse
//	@Failure		500			{object}	ErrResponse
//	@Router			/session/{session_id}/decrypt   [post]
func (h *Handlers) createDecrypt(w http.ResponseWriter, r *http.Request) {
	data := &DecryptRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	s := r.Context().Value("session").(*sessionstore.Session)
	plaintext, err := encryption.Decrypt(
		algorithmFromText(s.AlgorithmName),
		[]byte(s.Key),
		data.Ciphertext,
	)
	if err != nil {
		// FIXME: error is wrong
		h.logger.Info("error here")
		render.Render(w, r, h.ErrInternalServer(err))
		return
	}

	render.Status(r, http.StatusOK)
	h.logger.Info("finally here")
	render.Render(w, r, &DecryptResponse{Plaintext: plaintext})
}

// Encrypt a non-encoded plaintext input and returns a base64 encoded cipher text.
//
//	@Summary		Encrypt plaintext.
//	@Description	Encrypt plaintext in the context of a specific encryption session.
//	@Description	The plaintext will be encrypted using the specific algorithm and key associated with the session.
//	@Tags			encryption, session
//	@Accept			json
//	@Produce		json
//	@Param			session_id	path		string			false	"An encryption session ID"
//	@Param			request		body		EncryptRequest	true	"Request body"
//	@Success		200			{object}	EncryptResponse
//	@Failure		400			{object}	ErrResponse
//	@Failure		404			{object}	ErrResponse
//	@Failure		500			{object}	ErrResponse
//	@Router			/session/{session_id}/encrypt   [post]
func (h *Handlers) createEncrypt(w http.ResponseWriter, r *http.Request) {
	data := &EncryptRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	s := r.Context().Value("session").(*sessionstore.Session)
	cipherText, err := encryption.Encrypt(
		algorithmFromText(s.AlgorithmName),
		[]byte(s.Key),
		data.Plaintext,
	)
	if err != nil {
		render.Render(w, r, h.ErrInternalServer(err))
		return
	}

	render.Status(r, http.StatusOK)
	render.Render(w, r, EncryptResponse{CipherText: cipherText})
}

// Creates an encryption session given an algorithm type and key.
//
//	@Summary		Create encryption session.
//	@Description	Create an encryption session associating a session with a specific algorithm and key.
//	@Tags			encryption, session
//	@Accept			json
//	@Produce		json
//	@Param			request	body		SessionRequest	true	"Request body"
//	@Success		200		{object}	SessionResponse
//	@Failure		400		{object}	ErrResponse
//	@Failure		500		{object}	ErrResponse
//	@Router			/session   [post]
func (h *Handlers) createSession(w http.ResponseWriter, r *http.Request) {
	data := &SessionRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	id, err := h.sessionStore.NewSession(data.AlgorithmName, data.Key)
	if err != nil {
		render.Render(w, r, h.ErrInternalServer(err))
		return
	}

	render.Status(r, http.StatusCreated)
	render.Render(w, r, &SessionResponse{ID: id})
}

// Retrieves the list of supported symmetric encryption algorithms.
//
//	@Summary		List supported symmetric encryption algorithms.
//	@Description	Returns a list of all supported symmetric encryption algorithms.
//	@Description	These can then be used when creating a session.
//	@Tags			encryption, algorithms
//	@Produce		json
//	@Success		200	{object}	AlgorithmsResponse
//	@Router			/algorithms   [get]
func (h *Handlers) getAlgorithms(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusOK)
	render.Render(w, r, &AlgorithmsResponse{Names: encryption.Algorithms()})
}
