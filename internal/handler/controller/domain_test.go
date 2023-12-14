package controller_test

import (
	"bytes"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oursky/pageship/internal/api"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/handler/controller"
	"github.com/oursky/pageship/internal/models"
	"github.com/oursky/pageship/testutil"
	"github.com/stretchr/testify/assert"
)

var defaultConfig = controller.Config{
	TokenSigningKey: bytes.NewBufferString("test").Bytes(),
	TokenAuthority:  "test",
}

func TestDomainListDomains(t *testing.T) {
	t.Run("Should list all of domains", func(t *testing.T) {
		testutil.WithTestController(func(c *testutil.TestController) {
			user, token := c.SigninUser("mock user")
			c.NewApp("test", user, nil)
			req := httptest.NewRequest("GET", "http://localtest.me/api/v1/apps/test/domains", nil)
			req.Header.Add("Authorization", "bearer "+token)
			w := httptest.NewRecorder()
			c.ServeHTTP(w, req)
			domains, err := testutil.DecodeJSONResponse[[]*api.APIDomain](w.Result())
			if assert.NoError(t, err) {
				assert.Equal(t, 0, len(domains))
			}

		})
	})
}

func TestDomainCreation(t *testing.T) {
	t.Run("Should raise domain is undefined", func(t *testing.T) {
		testutil.WithTestController(func(c *testutil.TestController) {
			user, token := c.SigninUser("mock user")
			c.NewApp("test", user, nil)
			req := httptest.NewRequest("POST", "http://localtest.me/api/v1/apps/test/domains/test-domain", nil)
			req.Header.Add("Authorization", "bearer "+token)
			w := httptest.NewRecorder()
			c.ServeHTTP(w, req)
			_, err := testutil.DecodeJSONResponse[*api.APIDomain](w.Result())
			if assert.Error(t, err) {
				err := err.(api.ServerError)
				assert.Equal(t, 400, err.Code)
				assert.Equal(t, models.ErrUndefinedDomain.Error(), err.Message)
			}
		})
	})
	t.Run("Should create new domain", func(t *testing.T) {
		testutil.WithTestController(func(c *testutil.TestController) {
			user, token := c.SigninUser("mock user")
			c.NewApp("test", user, &config.AppConfig{
				Domains: []config.AppDomainConfig{
					{
						Site:   "main",
						Domain: "test.com",
					},
				},
				Team: []*config.AccessRule{
					{

						ACLSubjectRule: config.ACLSubjectRule{
							PageshipUser: user.ID,
						},
						Access: config.AccessLevelAdmin,
					},
				},
			})
			req := httptest.NewRequest("POST", "http://localtest.me/api/v1/apps/test/domains/test.com", nil)
			req.Header.Add("Authorization", "bearer "+token)
			w := httptest.NewRecorder()
			c.ServeHTTP(w, req)
			domain, err := testutil.DecodeJSONResponse[*api.APIDomain](w.Result())
			if assert.NoError(t, err) {
				assert.Equal(t, "test.com", domain.Domain.Domain)
			}
		})
	})
	t.Run("Duplicate domain", func(t *testing.T) {
		testutil.WithTestController(func(c *testutil.TestController) {
			user, token := c.SigninUser("mock user")
			c.NewApp("test", user, &config.AppConfig{
				Domains: []config.AppDomainConfig{
					{
						Site:   "main",
						Domain: "test.com",
					},
				},
				Team: []*config.AccessRule{
					{

						ACLSubjectRule: config.ACLSubjectRule{
							PageshipUser: user.ID,
						},
						Access: config.AccessLevelAdmin,
					},
				},
			})
			c.NewApp("test2", user, &config.AppConfig{
				Domains: []config.AppDomainConfig{
					{
						Site:   "main",
						Domain: "test.com",
					},
				},
				Team: []*config.AccessRule{
					{

						ACLSubjectRule: config.ACLSubjectRule{
							PageshipUser: user.ID,
						},
						Access: config.AccessLevelAdmin,
					},
				},
			})
			req := httptest.NewRequest("POST", "http://localtest.me/api/v1/apps/test/domains/test.com", nil)
			req.Header.Add("Authorization", "bearer "+token)
			w := httptest.NewRecorder()
			c.ServeHTTP(w, req)
			domain, err := testutil.DecodeJSONResponse[*api.APIDomain](w.Result())
			if assert.NoError(t, err) {
				assert.Equal(t, "test.com", domain.Domain.Domain)
			}
			t.Run("Should raise used domain error", func(t *testing.T) {
				req = httptest.NewRequest("POST", "http://localtest.me/api/v1/apps/test2/domains/test.com", nil)
				req.Header.Add("Authorization", "bearer "+token)
				w = httptest.NewRecorder()
				c.ServeHTTP(w, req)
				domain, err = testutil.DecodeJSONResponse[*api.APIDomain](w.Result())
				if assert.Error(t, err) {
					assert.Equal(t, 409, err.(api.ServerError).Code)
					assert.Equal(t, models.ErrDomainUsedName.Error(), err.(api.ServerError).Message)
				}
			})
			t.Run("Should replace app of the used domain", func(t *testing.T) {
				req := httptest.NewRequest("POST", "http://localtest.me/api/v1/apps/test2/domains/test.com?replaceApp=test", nil)
				req.Header.Add("Authorization", "bearer "+token)
				w = httptest.NewRecorder()
				c.ServeHTTP(w, req)
				domain, err = testutil.DecodeJSONResponse[*api.APIDomain](w.Result())
				if assert.NoError(t, err) {
					assert.Equal(t, "test2", domain.Domain.AppID)
				}
			})
		})
	})
	t.Run("Should get domain when it exists", func(t *testing.T) {
		testutil.WithTestController(func(c *testutil.TestController) {
			user, token := c.SigninUser("mock user")
			c.NewApp("test", user, &config.AppConfig{
				Domains: []config.AppDomainConfig{
					{
						Site:   "main",
						Domain: "test.com",
					},
				},
				Team: []*config.AccessRule{
					{

						ACLSubjectRule: config.ACLSubjectRule{
							PageshipUser: user.ID,
						},
						Access: config.AccessLevelAdmin,
					},
				},
			})
			req := httptest.NewRequest("POST", "http://localtest.me/api/v1/apps/test/domains/test.com", nil)
			req.Header.Add("Authorization", "bearer "+token)
			w := httptest.NewRecorder()
			c.ServeHTTP(w, req)
			domain, err := testutil.DecodeJSONResponse[*api.APIDomain](w.Result())
			if assert.NoError(t, err) {
				assert.Equal(t, "test.com", domain.Domain.Domain)
			}
			w = httptest.NewRecorder()
			c.ServeHTTP(w, req)
			domain, err = testutil.DecodeJSONResponse[*api.APIDomain](w.Result())
			if assert.NoError(t, err) {
				assert.Equal(t, "test.com", domain.Domain.Domain)
			}
		})
	})
}

func TestDomainDeletion(t *testing.T) {
	t.Run("Should raise domain not defined", func(t *testing.T) {
		testutil.WithTestController(func(c *testutil.TestController) {
			user, token := c.SigninUser("mock user")
			c.NewApp("test", user, &config.AppConfig{
				Domains: []config.AppDomainConfig{
					{
						Site:   "main",
						Domain: "test.com",
					},
				},
				Team: []*config.AccessRule{
					{

						ACLSubjectRule: config.ACLSubjectRule{
							PageshipUser: user.ID,
						},
						Access: config.AccessLevelAdmin,
					},
				},
			})
			req := httptest.NewRequest("DELETE", "http://localtest.me/api/v1/apps/test/domains/test.com", nil)
			req.Header.Add("Authorization", "bearer "+token)
			w := httptest.NewRecorder()
			c.ServeHTTP(w, req)
			_, err := testutil.DecodeJSONResponse[*api.APIDomain](w.Result())
			if assert.Error(t, err) {
				assert.Equal(t, 404, err.(api.ServerError).Code)
				assert.Equal(t, models.ErrDomainNotFound.Error(), err.(api.ServerError).Message)
			}
		})
	})
	t.Run("Should delete domain", func(t *testing.T) {
		testutil.WithTestController(func(c *testutil.TestController) {
			user, token := c.SigninUser("mock user")
			c.NewApp("test", user, &config.AppConfig{
				Domains: []config.AppDomainConfig{
					{
						Site:   "main",
						Domain: "test.com",
					},
				},
				Team: []*config.AccessRule{
					{

						ACLSubjectRule: config.ACLSubjectRule{
							PageshipUser: user.ID,
						},
						Access: config.AccessLevelAdmin,
					},
				},
			})
			req := httptest.NewRequest("POST", "http://localtest.me/api/v1/apps/test/domains/test.com", nil)
			req.Header.Add("Authorization", "bearer "+token)
			w := httptest.NewRecorder()
			c.ServeHTTP(w, req)
			domain, err := testutil.DecodeJSONResponse[*api.APIDomain](w.Result())
			if assert.NoError(t, err) {
				assert.Equal(t, "test.com", domain.Domain.Domain)
			}

			req = httptest.NewRequest("DELETE", "http://localtest.me/api/v1/apps/test/domains/test.com", nil)
			req.Header.Add("Authorization", "bearer "+token)
			w = httptest.NewRecorder()
			c.ServeHTTP(w, req)
			deleteResult := w.Result()
			assert.Equal(t, 200, deleteResult.StatusCode)
			_, err = c.DB.GetDomainByName(c.Context, "test.com")
			if assert.Error(t, err) {
				assert.ErrorIs(t, models.ErrDomainNotFound, err)
			}
		})
	})
}

func setupDomainVerification(c *testutil.TestController) (token string) {

	c.UpdateConfig(func(config *controller.Config) {
		config.DomainVerificationEnabled = true
	})
	user, token := c.SigninUser("mock user")
	c.NewApp("test", user, &config.AppConfig{
		Domains: []config.AppDomainConfig{
			{
				Site:   "main",
				Domain: "test.com",
			},
		},
		Team: []*config.AccessRule{
			{

				ACLSubjectRule: config.ACLSubjectRule{
					PageshipUser: user.ID,
				},
				Access: config.AccessLevelAdmin,
			},
		},
	})
	c.NewApp("test2", user, &config.AppConfig{
		Domains: []config.AppDomainConfig{
			{
				Site:   "main",
				Domain: "test.com",
			},
		},
		Team: []*config.AccessRule{
			{

				ACLSubjectRule: config.ACLSubjectRule{
					PageshipUser: user.ID,
				},
				Access: config.AccessLevelAdmin,
			},
		},
	})
	return
}

func TestDomainVerification(t *testing.T) {
	t.Run("Should raise domain is undefined", func(t *testing.T) {
		testutil.WithTestController(func(c *testutil.TestController) {
			token := setupDomainVerification(c)
			req := httptest.NewRequest("POST", "http://localtest.me/api/v1/apps/test/domains/test-undefined.com", nil)
			req.Header.Add("Authorization", "bearer "+token)
			w := httptest.NewRecorder()
			c.ServeHTTP(w, req)
			_, err := testutil.DecodeJSONResponse[*api.APIDomain](w.Result())
			if assert.Error(t, err) {
				assert.Equal(t, models.ErrUndefinedDomain.Error(), err.(api.ServerError).Message)
			}
		})
	})
	t.Run("Should add a pending activating domain for same domain with different Apps", func(t *testing.T) {
		testutil.WithTestController(func(c *testutil.TestController) {
			token := setupDomainVerification(c)

			t.Run("Should add a pending activating domain with app B", func(t *testing.T) {
				req := httptest.NewRequest("POST", "http://localtest.me/api/v1/apps/test/domains/test.com", nil)
				req.Header.Add("Authorization", "bearer "+token)
				w := httptest.NewRecorder()
				c.ServeHTTP(w, req)
				domain, err := testutil.DecodeJSONResponse[*api.APIDomain](w.Result())
				if assert.NoError(t, err) {
					assert.Nil(t, domain.Domain)
					assert.Equal(t, "test.com", domain.DomainVerification.Domain)
					assert.Equal(t, "test", domain.DomainVerification.AppID)
				}
			})
			t.Run("Should add a pending activating domain with app B", func(t *testing.T) {
				req := httptest.NewRequest("POST", "http://localtest.me/api/v1/apps/test2/domains/test.com", nil)
				req.Header.Add("Authorization", "bearer "+token)
				w := httptest.NewRecorder()
				c.ServeHTTP(w, req)
				domain, err := testutil.DecodeJSONResponse[*api.APIDomain](w.Result())
				if assert.NoError(t, err) {
					assert.Nil(t, domain.Domain)
					assert.Equal(t, "test.com", domain.DomainVerification.Domain)
					assert.Equal(t, "test2", domain.DomainVerification.AppID)
				}
			})
		})
	})
			req := httptest.NewRequest("POST", "http://localtest.me/api/v1/apps/test/domains/test.com", nil)
			req.Header.Add("Authorization", "bearer "+token)
			w := httptest.NewRecorder()
			c.ServeHTTP(w, req)
			domain, err := testutil.DecodeJSONResponse[*api.APIDomain](w.Result())
			if assert.NoError(t, err) {
				assert.Nil(t, domain.Domain)
				assert.Equal(t, "test.com", domain.DomainVerification.Domain)
			}
		})
	})
	testutil.WithTestController(func(c *testutil.TestController) {
		token := setupDomainVerification(c)
		db.WithTx(c.Context, c.DB, func(tx db.Tx) error {
			return tx.CreateDomain(c.Context, models.NewDomain(time.Now(), "test.com", "test", "main"))
		})

		t.Run("Should raise used domain error", func(t *testing.T) {
			req := httptest.NewRequest("POST", "http://localtest.me/api/v1/apps/test2/domains/test.com", nil)
			req.Header.Add("Authorization", "bearer "+token)
			w := httptest.NewRecorder()
			c.ServeHTTP(w, req)
			_, err := testutil.DecodeJSONResponse[*api.APIDomain](w.Result())
			if assert.Error(t, err) {
				assert.Equal(t, 409, err.(api.ServerError).Code)
				assert.Equal(t, models.ErrDomainUsedName.Error(), err.(api.ServerError).Message)
			}
		})

		t.Run("Should replace app of the used domain", func(t *testing.T) {
			req := httptest.NewRequest("POST", "http://localtest.me/api/v1/apps/test2/domains/test.com?replaceApp=test", nil)
			req.Header.Add("Authorization", "bearer "+token)
			w := httptest.NewRecorder()
			c.ServeHTTP(w, req)
			domain, err := testutil.DecodeJSONResponse[*api.APIDomain](w.Result())
			if assert.NoError(t, err) {
				assert.Nil(t, domain.DomainVerification)
				assert.NotNil(t, domain.Domain)
				assert.Equal(t, "test.com", domain.Domain.Domain)
				assert.Equal(t, "test2", domain.Domain.AppID)
			}
		})
	})
}
