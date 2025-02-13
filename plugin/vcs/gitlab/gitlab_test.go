package gitlab

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/youzi-1122/bytebase/common"
	"github.com/youzi-1122/bytebase/plugin/vcs"
	"github.com/youzi-1122/bytebase/plugin/vcs/internal/oauth"
)

func TestProvider_FetchUserInfo(t *testing.T) {
	p := newProvider(
		vcs.ProviderConfig{
			Client: &http.Client{
				Transport: &common.MockRoundTripper{
					MockRoundTrip: func(r *http.Request) (*http.Response, error) {
						assert.Equal(t, "/api/v4/users/1", r.URL.Path)
						return &http.Response{
							StatusCode: http.StatusOK,
							// Example response taken from https://docs.gitlab.com/ee/api/users.html#single-user
							Body: io.NopCloser(strings.NewReader(`
{
  "id": 1,
  "username": "john_smith",
  "name": "John Smith",
  "state": "active",
  "avatar_url": "http://localhost:3000/uploads/user/avatar/1/cd8.jpeg",
  "web_url": "http://localhost:3000/john_smith",
  "created_at": "2012-05-23T08:00:58Z",
  "bio": "",
  "bot": false,
  "location": null,
  "public_email": "john@example.com",
  "skype": "",
  "linkedin": "",
  "twitter": "",
  "website_url": "",
  "organization": "",
  "job_title": "Operations Specialist",
  "followers": 1,
  "following": 1
}
`)),
						}, nil
					},
				},
			},
		},
	)

	ctx := context.Background()
	got, err := p.FetchUserInfo(ctx, common.OauthContext{}, "", "1")
	require.NoError(t, err)

	want := &vcs.UserInfo{
		PublicEmail: "john@example.com",
		Name:        "John Smith",
		State:       vcs.StateActive,
	}
	assert.Equal(t, want, got)
}

func TestProvider_FetchRepositoryActiveMemberList(t *testing.T) {
	t.Run("missing public email", func(t *testing.T) {
		p := newProvider(
			vcs.ProviderConfig{
				Client: &http.Client{
					Transport: &common.MockRoundTripper{
						MockRoundTrip: func(r *http.Request) (*http.Response, error) {
							switch r.URL.Path {
							case "/api/v4/projects/1/members/all":
								return &http.Response{
									StatusCode: http.StatusOK,
									// Example response derived from https://docs.gitlab.com/ee/api/members.html#list-all-members-of-a-group-or-project-including-inherited-and-invited-members
									Body: io.NopCloser(strings.NewReader(`
[
  {
    "id": 1,
    "username": "raymond_smith",
    "name": "Raymond Smith",
    "state": "active",
    "avatar_url": "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
    "web_url": "http://192.168.1.8:3000/root",
    "created_at": "2012-09-22T14:13:35Z",
    "created_by": {
      "id": 2,
      "username": "john_doe",
      "name": "John Doe",
      "state": "active",
      "avatar_url": "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
      "web_url": "http://192.168.1.8:3000/root"
    },
    "expires_at": "2012-10-22T14:13:35Z",
    "access_level": 30,
    "group_saml_identity": null,
    "membership_state": "active"
  }
]
`)),
								}, nil
							case "/api/v4/users/1":
								return &http.Response{
									StatusCode: http.StatusOK,
									// Example response derived from https://docs.gitlab.com/ee/api/users.html#single-user
									Body: io.NopCloser(strings.NewReader(`
{
  "id": 1,
  "username": "john_smith",
  "name": "John Smith",
  "state": "active",
  "avatar_url": "http://localhost:3000/uploads/user/avatar/1/cd8.jpeg",
  "web_url": "http://localhost:3000/john_smith",
  "created_at": "2012-05-23T08:00:58Z",
  "bio": "",
  "bot": false,
  "location": null,
  "public_email": "",
  "skype": "",
  "linkedin": "",
  "twitter": "",
  "website_url": "",
  "organization": "",
  "job_title": "Operations Specialist",
  "followers": 1,
  "following": 1
}
`)),
								}, nil
							}
							return nil, errors.Errorf("unexpected request path: %s", r.URL.Path)
						},
					},
				},
			},
		)

		ctx := context.Background()
		_, got := p.FetchRepositoryActiveMemberList(ctx, common.OauthContext{}, "", "1")
		want := "[ Raymond Smith ] did not configure their public email in GitLab, please make sure every members' public email is configured before syncing, see https://docs.gitlab.com/ee/user/profile"
		assert.EqualError(t, got, want)
	})

	p := newProvider(
		vcs.ProviderConfig{
			Client: &http.Client{
				Transport: &common.MockRoundTripper{
					MockRoundTrip: func(r *http.Request) (*http.Response, error) {
						switch r.URL.Path {
						case "/api/v4/projects/1/members/all":
							return &http.Response{
								StatusCode: http.StatusOK,
								// Example response derived from https://docs.gitlab.com/ee/api/members.html#list-all-members-of-a-group-or-project-including-inherited-and-invited-members
								Body: io.NopCloser(strings.NewReader(`
[
  {
    "id": 1,
    "username": "raymond_smith",
    "name": "Raymond Smith",
    "state": "active",
    "avatar_url": "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
    "web_url": "http://192.168.1.8:3000/root",
    "created_at": "2012-09-22T14:13:35Z",
    "created_by": {
      "id": 2,
      "username": "john_doe",
      "name": "John Doe",
      "state": "active",
      "avatar_url": "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
      "web_url": "http://192.168.1.8:3000/root"
    },
    "expires_at": "2012-10-22T14:13:35Z",
    "access_level": 30,
    "group_saml_identity": null,
    "membership_state": "active"
  },
  {
    "id": 2,
    "username": "john_doe",
    "name": "John Doe",
    "state": "archived",
    "avatar_url": "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
    "web_url": "http://192.168.1.8:3000/root",
    "created_at": "2012-09-22T14:13:35Z",
    "created_by": {
      "id": 1,
      "username": "raymond_smith",
      "name": "Raymond Smith",
      "state": "active",
      "avatar_url": "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
      "web_url": "http://192.168.1.8:3000/root"
    },
    "expires_at": "2012-10-22T14:13:35Z",
    "access_level": 30,
    "email": "john@example.com",
    "group_saml_identity": {
      "extern_uid":"ABC-1234567890",
      "provider": "group_saml",
      "saml_provider_id": 10
    },
    "membership_state": "active"
  }
]
`)),
							}, nil
						case "/api/v4/users/1":
							return &http.Response{
								StatusCode: http.StatusOK,
								// Example response taken from https://docs.gitlab.com/ee/api/users.html#single-user
								Body: io.NopCloser(strings.NewReader(`
{
  "id": 1,
  "username": "john_smith",
  "name": "John Smith",
  "state": "active",
  "avatar_url": "http://localhost:3000/uploads/user/avatar/1/cd8.jpeg",
  "web_url": "http://localhost:3000/john_smith",
  "created_at": "2012-05-23T08:00:58Z",
  "bio": "",
  "bot": false,
  "location": null,
  "public_email": "john@example.com",
  "skype": "",
  "linkedin": "",
  "twitter": "",
  "website_url": "",
  "organization": "",
  "job_title": "Operations Specialist",
  "followers": 1,
  "following": 1
}
`)),
							}, nil
						}
						return nil, errors.Errorf("unexpected request path: %s", r.URL.Path)
					},
				},
			},
		},
	)

	ctx := context.Background()
	got, err := p.FetchRepositoryActiveMemberList(ctx, common.OauthContext{}, "", "1")
	require.NoError(t, err)

	want := []*vcs.RepositoryMember{
		{
			Email:        "john@example.com",
			Name:         "Raymond Smith",
			State:        vcs.StateActive,
			Role:         common.ProjectDeveloper,
			VCSRole:      string(ProjectRoleDeveloper),
			RoleProvider: vcs.GitLabSelfHost,
		},
	}
	assert.Equal(t, want, got)
}

func TestProvider_FetchCommitByID(t *testing.T) {
	p := newProvider(
		vcs.ProviderConfig{
			Client: &http.Client{
				Transport: &common.MockRoundTripper{
					MockRoundTrip: func(r *http.Request) (*http.Response, error) {
						assert.Equal(t, "/api/v4/projects/5/repository/commits/6104942438c14ec7bd21c6cd5bd995272b3faff6", r.URL.Path)
						return &http.Response{
							StatusCode: http.StatusOK,
							// Example response taken from https://docs.gitlab.com/ee/api/commits.html#get-a-single-commit
							Body: io.NopCloser(strings.NewReader(`
{
  "id": "6104942438c14ec7bd21c6cd5bd995272b3faff6",
  "short_id": "6104942438c",
  "title": "Sanitize for network graph",
  "author_name": "randx",
  "author_email": "user@example.com",
  "committer_name": "Dmitriy",
  "committer_email": "user@example.com",
  "created_at": "2021-09-20T09:06:12.300+03:00",
  "message": "Sanitize for network graph",
  "committed_date": "2021-09-20T09:06:12.300+03:00",
  "authored_date": "2021-09-20T09:06:12.420+03:00",
  "parent_ids": [
    "ae1d9fb46aa2b07ee9836d49862ec4e2c46fbbba"
  ],
  "last_pipeline" : {
    "id": 8,
    "ref": "master",
    "sha": "2dc6aa325a317eda67812f05600bdf0fcdc70ab0",
    "status": "created"
  },
  "stats": {
    "additions": 15,
    "deletions": 10,
    "total": 25
  },
  "status": "running",
  "web_url": "https://gitlab.example.com/thedude/gitlab-foss/-/commit/6104942438c14ec7bd21c6cd5bd995272b3faff6"
}
`)),
						}, nil
					},
				},
			},
		},
	)

	ctx := context.Background()
	got, err := p.FetchCommitByID(ctx, common.OauthContext{}, "", "5", "6104942438c14ec7bd21c6cd5bd995272b3faff6")
	require.NoError(t, err)

	want := &vcs.Commit{
		ID:         "6104942438c14ec7bd21c6cd5bd995272b3faff6",
		AuthorName: "randx",
		CreatedTs:  1632117972,
	}
	assert.Equal(t, want, got)
}

func TestProvider_ExchangeOAuthToken(t *testing.T) {
	p := newProvider(
		vcs.ProviderConfig{
			Client: &http.Client{
				Transport: &common.MockRoundTripper{
					MockRoundTrip: func(r *http.Request) (*http.Response, error) {
						assert.Equal(t, "/oauth/token", r.URL.Path)
						assert.Equal(t, "client_id=test_client_id&client_secret=test_client_secret&code=test_code&grant_type=authorization_code&redirect_uri=http%3A%2F%2Flocalhost%3A3000", r.URL.RawQuery)
						return &http.Response{
							StatusCode: http.StatusOK,
							// Example response taken from https://docs.gitlab.com/ee/api/oauth2.html#authorization-code-flow
							Body: io.NopCloser(strings.NewReader(`
{
 "access_token": "de6780bc506a0446309bd9362820ba8aed28aa506c71eedbe1c5c4f9dd350e54",
 "token_type": "bearer",
 "expires_in": 7200,
 "refresh_token": "8257e65c97202ed1726cf9571600918f3bffb2544b26e00a61df9897668c33a1",
 "created_at": 1607635748
}
`)),
						}, nil
					},
				},
			},
		},
	)

	ctx := context.Background()
	got, err := p.ExchangeOAuthToken(ctx, "",
		&common.OAuthExchange{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
			Code:         "test_code",
			RedirectURL:  "http://localhost:3000",
		},
	)
	require.NoError(t, err)

	want := &vcs.OAuthToken{
		AccessToken:  "de6780bc506a0446309bd9362820ba8aed28aa506c71eedbe1c5c4f9dd350e54",
		RefreshToken: "8257e65c97202ed1726cf9571600918f3bffb2544b26e00a61df9897668c33a1",
		ExpiresIn:    7200,
		CreatedAt:    1607635748,
		ExpiresTs:    1607642948,
	}
	assert.Equal(t, want, got)
}

func TestProvider_FetchAllRepositoryList(t *testing.T) {
	p := newProvider(
		vcs.ProviderConfig{
			Client: &http.Client{
				Transport: &common.MockRoundTripper{
					MockRoundTrip: func(r *http.Request) (*http.Response, error) {
						assert.Equal(t, "/api/v4/projects", r.URL.Path)
						return &http.Response{
							StatusCode: http.StatusOK,
							// Example response taken from https://docs.gitlab.com/ee/api/projects.html#list-all-projects
							Body: io.NopCloser(strings.NewReader(`
[
  {
    "id": 4,
    "description": null,
    "default_branch": "master",
    "ssh_url_to_repo": "git@example.com:diaspora/diaspora-client.git",
    "http_url_to_repo": "http://example.com/diaspora/diaspora-client.git",
    "web_url": "http://example.com/diaspora/diaspora-client",
    "readme_url": "http://example.com/diaspora/diaspora-client/blob/master/README.md",
    "tag_list": [
      "example",
      "disapora client"
    ],
    "topics": [
      "example",
      "disapora client"
    ],
    "name": "Diaspora Client",
    "name_with_namespace": "Diaspora / Diaspora Client",
    "path": "diaspora-client",
    "path_with_namespace": "diaspora/diaspora-client",
    "created_at": "2013-09-30T13:46:02Z",
    "last_activity_at": "2013-09-30T13:46:02Z",
    "forks_count": 0,
    "avatar_url": "http://example.com/uploads/project/avatar/4/uploads/avatar.png",
    "star_count": 0
  }
]
`)),
						}, nil
					},
				},
			},
		},
	)

	ctx := context.Background()
	got, err := p.FetchAllRepositoryList(ctx, common.OauthContext{}, "")
	require.NoError(t, err)

	// Non-active member should be excluded
	want := []*vcs.Repository{
		{
			ID:       4,
			Name:     "Diaspora Client",
			FullPath: "diaspora/diaspora-client",
			WebURL:   "http://example.com/diaspora/diaspora-client",
		},
	}
	assert.Equal(t, want, got)
}

func TestOAuth_RefreshToken(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{
		Transport: &common.MockRoundTripper{
			MockRoundTrip: func(r *http.Request) (*http.Response, error) {
				token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
				if token == "expired" {
					return &http.Response{
						StatusCode: http.StatusBadRequest,
						Body: io.NopCloser(strings.NewReader(`
					{"error":"invalid_token","error_description":"Token is expired. You can either do re-authorization or token refresh."}
					`)),
					}, nil
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					// Example response taken from https://docs.gitlab.com/ee/api/oauth2.html#authorization-code-with-proof-key-for-code-exchange-pkce
					Body: io.NopCloser(strings.NewReader(`
{
 "access_token": "de6780bc506a0446309bd9362820ba8aed28aa506c71eedbe1c5c4f9dd350e54",
 "token_type": "bearer",
 "expires_in": 7200,
 "refresh_token": "8257e65c97202ed1726cf9571600918f3bffb2544b26e00a61df9897668c33a1",
 "created_at": 1607635748
}
`)),
				}, nil
			},
		},
	}
	token := "expired"

	calledRefresher := false
	refresher := func(_, _ string, _ int64) error {
		calledRefresher = true
		return nil
	}

	_, _, err := oauth.Get(
		ctx,
		client,
		"https://gitlab.example.com/api/v4/users/octocat",
		&token,
		tokenRefresher(
			"https://gitlab.example.com",
			oauthContext{},
			refresher,
		),
	)
	require.NoError(t, err)
	assert.Equal(t, "de6780bc506a0446309bd9362820ba8aed28aa506c71eedbe1c5c4f9dd350e54", token)
	assert.True(t, calledRefresher)
}
