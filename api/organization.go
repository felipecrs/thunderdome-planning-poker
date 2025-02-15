package api

import (
	"net/http"
	"strings"

	"github.com/StevenWeathers/thunderdome-planning-poker/model"
	"github.com/gorilla/mux"
)

type organizationResponse struct {
	Organization *model.Organization `json:"organization"`
	Role         string              `json:"role"`
}

type orgTeamResponse struct {
	Organization     *model.Organization `json:"organization"`
	Team             *model.Team         `json:"team"`
	OrganizationRole string              `json:"organizationRole"`
	TeamRole         string              `json:"teamRole"`
}

// handleGetOrganizationsByUser gets a list of organizations the user is a part of
// @Summary Get Users Organizations
// @Description get list of organizations for the authenticated user
// @Tags organization
// @Produce  json
// @Param userId path string true "the user ID to get organizations for"
// @Param limit query int false "Max number of results to return"
// @Param offset query int false "Starting point to return rows from, should be multiplied by limit or 0"
// @Success 200 object standardJsonResponse{data=[]model.Organization}
// @Failure 403 object standardJsonResponse{}
// @Security ApiKeyAuth
// @Router /users/{userId}/organizations [get]
func (a *api) handleGetOrganizationsByUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		UserID := vars["userId"]

		Limit, Offset := getLimitOffsetFromRequest(r, w)

		Organizations := a.db.OrganizationListByUser(UserID, Limit, Offset)

		Success(w, r, http.StatusOK, Organizations, nil)
	}
}

// handleGetOrganizationByUser gets an organization with user role
// @Summary Get Organization
// @Description get an organization with user role
// @Tags organization
// @Produce  json
// @Param orgId path string true "organization id"
// @Success 200 object standardJsonResponse{data=organizationResponse}
// @Failure 403 object standardJsonResponse{}
// @Failure 500 object standardJsonResponse{}
// @Security ApiKeyAuth
// @Router /organizations/{orgId} [get]
func (a *api) handleGetOrganizationByUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		OrgRole := r.Context().Value(contextKeyOrgRole).(string)
		vars := mux.Vars(r)
		OrgID := vars["orgId"]

		Organization, err := a.db.OrganizationGet(OrgID)
		if err != nil {
			Failure(w, r, http.StatusInternalServerError, err)
			return
		}

		result := &organizationResponse{
			Organization: Organization,
			Role:         OrgRole,
		}

		Success(w, r, http.StatusOK, result, nil)
	}
}

// handleCreateOrganization handles creating an organization with current user as admin
// @Summary Create Organization
// @Description Create organization with current user as admin
// @Tags organization
// @Produce  json
// @Param userId path string true "user id"
// @Param name body string true "the organization name"
// @Success 200 object standardJsonResponse{data=model.Organization}
// @Failure 403 object standardJsonResponse{}
// @Failure 500 object standardJsonResponse{}
// @Security ApiKeyAuth
// @Router /users/{userId}/organizations [post]
func (a *api) handleCreateOrganization() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		UserID := vars["userId"]

		keyVal := getJSONRequestBody(r, w)

		OrgName := keyVal["name"].(string)
		Organization, err := a.db.OrganizationCreate(UserID, OrgName)
		if err != nil {
			Failure(w, r, http.StatusInternalServerError, err)
			return
		}

		Success(w, r, http.StatusOK, Organization, nil)
	}
}

// handleGetOrganizationTeams gets a list of teams associated to the organization
// @Summary Get Organization Teams
// @Description get a list of organization teams
// @Tags organization
// @Produce  json
// @Param orgId path string true "organization id"
// @Success 200 object standardJsonResponse{data=[]model.Team}
// @Failure 403 object standardJsonResponse{}
// @Security ApiKeyAuth
// @Router /organizations/{orgId}/teams [get]
func (a *api) handleGetOrganizationTeams() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		OrgID := vars["orgId"]
		Limit, Offset := getLimitOffsetFromRequest(r, w)

		Teams := a.db.OrganizationTeamList(OrgID, Limit, Offset)

		Success(w, r, http.StatusOK, Teams, nil)
	}
}

// handleGetOrganizationUsers gets a list of users associated to the organization
// @Summary Get Organization Users
// @Description get a list of organization users
// @Tags organization
// @Produce  json
// @Param orgId path string true "organization id"
// @Success 200 object standardJsonResponse{data=[]model.User}
// @Failure 403 object standardJsonResponse{}
// @Security ApiKeyAuth
// @Router /organizations/{orgId}/users [get]
func (a *api) handleGetOrganizationUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		OrgID := vars["orgId"]
		Limit, Offset := getLimitOffsetFromRequest(r, w)

		Teams := a.db.OrganizationUserList(OrgID, Limit, Offset)

		Success(w, r, http.StatusOK, Teams, nil)
	}
}

// handleCreateOrganizationTeam handles creating an organization team
// @Summary Create Organization Team
// @Description Create organization team with current user as admin
// @Tags organization
// @Produce  json
// @Param orgId path string true "organization id"
// @Param name body string true "team name"
// @Success 200 object standardJsonResponse{data=model.Team}
// @Failure 403 object standardJsonResponse{}
// @Failure 500 object standardJsonResponse{}
// @Security ApiKeyAuth
// @Router /organizations/{orgId}/teams [post]
func (a *api) handleCreateOrganizationTeam() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		keyVal := getJSONRequestBody(r, w)

		TeamName := keyVal["name"].(string)
		OrgID := vars["orgId"]
		NewTeam, err := a.db.OrganizationTeamCreate(OrgID, TeamName)
		if err != nil {
			Failure(w, r, http.StatusInternalServerError, err)
			return
		}

		Success(w, r, http.StatusOK, NewTeam, nil)
	}
}

// handleOrganizationAddUser handles adding user to an organization
// @Summary Add Org User
// @Description Add user to organization
// @Tags organization
// @Produce  json
// @Param orgId path string true "organization id"
// @Param email body string true "the users email"
// @Param role body string true "the user's organization role" Enums(MEMBER, ADMIN)
// @Success 200 object standardJsonResponse{}
// @Failure 403 object standardJsonResponse{}
// @Failure 500 object standardJsonResponse{}
// @Security ApiKeyAuth
// @Router /organizations/{orgId}/users [post]
func (a *api) handleOrganizationAddUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		keyVal := getJSONRequestBody(r, w)

		vars := mux.Vars(r)
		OrgID := vars["orgId"]
		UserEmail := strings.ToLower(keyVal["email"].(string))
		Role := keyVal["role"].(string)

		User, UserErr := a.db.GetUserByEmail(UserEmail)
		if UserErr != nil {
			Failure(w, r, http.StatusInternalServerError, Errorf(ENOTFOUND, "USER_NOT_FOUND"))
			return
		}

		_, err := a.db.OrganizationAddUser(OrgID, User.Id, Role)
		if err != nil {
			Failure(w, r, http.StatusInternalServerError, err)
			return
		}

		Success(w, r, http.StatusOK, nil, nil)
	}
}

// handleOrganizationRemoveUser handles removing user from an organization (including departments, teams)
// @Summary Remove Org User
// @Description Remove user from organization including departments and teams
// @Tags organization
// @Produce  json
// @Param orgId path string true "organization id"
// @Param userId path string true "user id"
// @Success 200 object standardJsonResponse{}
// @Failure 403 object standardJsonResponse{}
// @Failure 500 object standardJsonResponse{}
// @Security ApiKeyAuth
// @Router /organizations/{orgId}/users/{userId} [delete]
func (a *api) handleOrganizationRemoveUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		OrgID := vars["orgId"]
		UserID := vars["userId"]

		err := a.db.OrganizationRemoveUser(OrgID, UserID)
		if err != nil {
			Failure(w, r, http.StatusInternalServerError, err)
			return
		}

		Success(w, r, http.StatusOK, nil, nil)
	}
}

// handleGetOrganizationTeamByUser gets a team with users roles
// @Summary Get Organization Team
// @Description Get an organizations team with users roles
// @Tags organization
// @Produce  json
// @Param orgId path string true "organization id"
// @Param teamId path string true "team id"
// @Success 200 object standardJsonResponse{data=orgTeamResponse}
// @Failure 403 object standardJsonResponse{}
// @Failure 500 object standardJsonResponse{}
// @Security ApiKeyAuth
// @Router /organizations/{orgId}/teams/{teamId} [get]
func (a *api) handleGetOrganizationTeamByUser() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		OrgRole := r.Context().Value(contextKeyOrgRole).(string)
		TeamRole := r.Context().Value(contextKeyTeamRole).(string)
		vars := mux.Vars(r)
		OrgID := vars["orgId"]
		TeamID := vars["teamId"]

		Organization, err := a.db.OrganizationGet(OrgID)
		if err != nil {
			Failure(w, r, http.StatusInternalServerError, err)
			return
		}

		Team, err := a.db.TeamGet(TeamID)
		if err != nil {
			Failure(w, r, http.StatusInternalServerError, err)
			return
		}

		result := &orgTeamResponse{
			Organization:     Organization,
			Team:             Team,
			OrganizationRole: OrgRole,
			TeamRole:         TeamRole,
		}

		Success(w, r, http.StatusOK, result, nil)
	}
}

// handleOrganizationTeamAddUser handles adding user to a team so long as they are in the organization
// @Summary Add Org Team User
// @Description Add user to organization team as long as they are already in the organization
// @Tags organization
// @Produce  json
// @Param orgId path string true "organization id"
// @Param teamId path string true "team id"
// @Param email body string true "the users email"
// @Param role body string true "the users team role" Enums(MEMBER, ADMIN)
// @Success 200 object standardJsonResponse{}
// @Failure 403 object standardJsonResponse{}
// @Failure 500 object standardJsonResponse{}
// @Security ApiKeyAuth
// @Router /organizations/{orgId}/teams/{teamId}/users [post]
func (a *api) handleOrganizationTeamAddUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		keyVal := getJSONRequestBody(r, w)

		vars := mux.Vars(r)
		OrgID := vars["orgId"]
		TeamID := vars["teamId"]
		UserEmail := strings.ToLower(keyVal["email"].(string))
		Role := keyVal["role"].(string)

		User, UserErr := a.db.GetUserByEmail(UserEmail)
		if UserErr != nil {
			Failure(w, r, http.StatusInternalServerError, Errorf(ENOTFOUND, "USER_NOT_FOUND"))
			return
		}

		OrgRole, roleErr := a.db.OrganizationUserRole(User.Id, OrgID)
		if OrgRole == "" || roleErr != nil {
			Failure(w, r, http.StatusInternalServerError, Errorf(EUNAUTHORIZED, "ORGANIZATION_USER_REQUIRED"))
			return
		}

		_, err := a.db.TeamAddUser(TeamID, User.Id, Role)
		if err != nil {
			Failure(w, r, http.StatusInternalServerError, err)
			return
		}

		Success(w, r, http.StatusOK, nil, nil)
	}
}
