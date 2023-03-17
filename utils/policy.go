package utils

import (
	"encoding/json"
	"errors"

	aclType "github.com/bnb-chain/greenfield/x/permission/types"
)

// Effect - policy statement effect Allow or Deny.
type Effect string
type Action string

const (
	AllowEffect            = "Allow"
	DenyEffect             = "Deny"
	UpdateBucketInfoAction = "gnfd:UpdateBucketInfo"
	DeleteBucketAction     = "gnfd:DeleteBucket"
	CreateObjectAction     = "gnfd:CreateObject"
	DeleteObjectAction     = "gnfd:DeleteObject"
	CopyObjectAction       = "gnfd:CopyObject"
	GetObjectAction        = "gnfd:GetObject"
	ExecuteObjectAction    = "gnfd:ExecuteObject"
	ListObjectAction       = "gnfd:ListObject"
	UpdateGroupAction      = "gnfd:UpdateGroupMember"
	DeleteGroupAction      = "gnfd:DeleteGroup"
)

var SupportActionList = map[Action]struct{}{
	UpdateBucketInfoAction: {},
	DeleteBucketAction:     {},
	CreateObjectAction:     {},
	DeleteObjectAction:     {},
	CopyObjectAction:       {},
	GetObjectAction:        {},
	ExecuteObjectAction:    {},
	ListObjectAction:       {},
	UpdateGroupAction:      {},
	DeleteGroupAction:      {},
}

var SupportActionMap = map[Action]aclType.ActionType{
	UpdateBucketInfoAction: aclType.ACTION_UPDATE_BUCKET_INFO,
	DeleteBucketAction:     aclType.ACTION_DELETE_BUCKET,
	CreateObjectAction:     aclType.ACTION_CREATE_OBJECT,
	DeleteObjectAction:     aclType.ACTION_DELETE_OBJECT,
	CopyObjectAction:       aclType.ACTION_COPY_OBJECT,
	GetObjectAction:        aclType.ACTION_GET_OBJECT,
	ExecuteObjectAction:    aclType.ACTION_EXECUTE_OBJECT,
	ListObjectAction:       aclType.ACTION_LIST_OBJECT,
	UpdateGroupAction:      aclType.ACTION_UPDATE_GROUP_MEMBER,
	DeleteGroupAction:      aclType.ACTION_DELETE_GROUP,
}

// GnfdPolicy - bucket policy.
type GnfdPolicy struct {
	Statements []GnfdStatement `json:"GnfdStatement"`
}

type ActionSet map[Action]struct{}

// GnfdStatement - policy statement.
type GnfdStatement struct {
	Effect  Effect   `json:"Effect"`
	Actions []Action `json:"Action"`
}

// Validate - checks if GnfdPolicy is valid or not.
func (g GnfdPolicy) Validate() error {
	for _, statement := range g.Statements {
		if err := statement.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (g GnfdPolicy) MarshalJSON() ([]byte, error) {
	if err := g.Validate(); err != nil {
		return nil, err
	}
	type newPolicy GnfdPolicy
	return json.Marshal(newPolicy(g))
}

func (g *GnfdPolicy) UnMarshal(content []byte) error {
	type newPolicy GnfdPolicy
	var policy newPolicy

	if err := json.Unmarshal(content, &policy); err != nil {
		return err
	}

	gnfdPolicy := GnfdPolicy(policy)
	if err := gnfdPolicy.Validate(); err != nil {
		return err
	}

	*g = gnfdPolicy
	return nil
}

// NewStatement - creates new statement.
func NewStatement(effect Effect, actionSet []Action) GnfdStatement {
	return GnfdStatement{
		Effect:  effect,
		Actions: actionSet,
	}
}

// Validate - checks if GnfdStatement is valid or not.
func (s GnfdStatement) Validate() error {
	if !s.Effect.IsValid() {
		return errors.New("invalid Effect" + string(s.Effect))
	}

	for _, action := range s.Actions {
		if !action.IsValid() {
			return errors.New("invalid action:" + string(action))
		}
	}

	return nil
}

func (s GnfdStatement) MarshalJSON() ([]byte, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	type newStatement GnfdStatement
	return json.Marshal(newStatement(s))

}

func (s *GnfdStatement) UnmarshalJSON(content []byte) error {
	type newStatement GnfdStatement
	var decodeStatement newStatement

	if err := json.Unmarshal(content, &decodeStatement); err != nil {
		return err
	}

	gnfdStatement := GnfdStatement(decodeStatement)
	if err := gnfdStatement.Validate(); err != nil {
		return err
	}

	*s = gnfdStatement
	return nil
}

// IsAllowed - returns if given check is allowed or not.
func (effect Effect) IsAllowed() bool {
	return effect == AllowEffect
}

// IsValid - checks if Effect is valid or not
func (effect Effect) IsValid() bool {
	switch effect {
	case AllowEffect, DenyEffect:
		return true
	}

	return false
}

// IsValid - checks if action is valid
func (action Action) IsValid() bool {
	_, ok := SupportActionList[action]
	return ok
}

func (action Action) MarshalJSON() ([]byte, error) {
	if action.IsValid() {
		return json.Marshal(string(action))
	}

	return nil, errors.New("invalid action" + string(action))
}

func (action *Action) UnmarshalJSON(content []byte) error {
	var decodeStr string

	if err := json.Unmarshal(content, &decodeStr); err != nil {
		return err
	}

	actionName := Action(decodeStr)
	if !actionName.IsValid() {
		return errors.New("invalid action :" + decodeStr)
	}

	*action = actionName
	return nil
}

func GetChainAction(action Action) aclType.ActionType {
	return SupportActionMap[action]
}

func GetChainEffect(effect Effect) aclType.Effect {
	if effect.IsAllowed() {
		return aclType.EFFECT_ALLOW
	} else {
		return aclType.EFFECT_DENY
	}
}

// ToSlice - returns slice of action set
func (a ActionSet) ToSlice() []Action {
	actions := []Action{}
	for action := range a {
		actions = append(actions, action)
	}
	return actions
}
