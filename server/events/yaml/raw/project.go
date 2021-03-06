package raw

import (
	"fmt"
	"strings"

	"github.com/go-ozzo/ozzo-validation"
	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
)

const (
	DefaultWorkspace         = "default"
	ApprovedApplyRequirement = "approved"
)

type Project struct {
	Name              *string   `yaml:"name,omitempty"`
	Dir               *string   `yaml:"dir,omitempty"`
	Workspace         *string   `yaml:"workspace,omitempty"`
	Workflow          *string   `yaml:"workflow,omitempty"`
	TerraformVersion  *string   `yaml:"terraform_version,omitempty"`
	Autoplan          *Autoplan `yaml:"autoplan,omitempty"`
	ApplyRequirements []string  `yaml:"apply_requirements,omitempty"`
}

func (p Project) Validate() error {
	hasDotDot := func(value interface{}) error {
		if strings.Contains(*value.(*string), "..") {
			return errors.New("cannot contain '..'")
		}
		return nil
	}
	validApplyReq := func(value interface{}) error {
		reqs := value.([]string)
		for _, r := range reqs {
			if r != ApprovedApplyRequirement {
				return fmt.Errorf("%q not supported, only %s is supported", r, ApprovedApplyRequirement)
			}
		}
		return nil
	}
	validTFVersion := func(value interface{}) error {
		strPtr := value.(*string)
		if strPtr == nil {
			return nil
		}
		_, err := version.NewVersion(*strPtr)
		return errors.Wrapf(err, "version %q could not be parsed", *strPtr)
	}
	return validation.ValidateStruct(&p,
		validation.Field(&p.Dir, validation.Required, validation.By(hasDotDot)),
		validation.Field(&p.ApplyRequirements, validation.By(validApplyReq)),
		validation.Field(&p.TerraformVersion, validation.By(validTFVersion)),
	)
}

func (p Project) ToValid() valid.Project {
	var v valid.Project
	v.Dir = *p.Dir

	if p.Workspace == nil {
		v.Workspace = DefaultWorkspace
	} else {
		v.Workspace = *p.Workspace
	}

	v.Workflow = p.Workflow
	if p.TerraformVersion != nil {
		v.TerraformVersion, _ = version.NewVersion(*p.TerraformVersion)
	}
	if p.Autoplan == nil {
		v.Autoplan = DefaultAutoPlan()
	} else {
		v.Autoplan = p.Autoplan.ToValid()
	}

	// There are no default apply requirements.
	v.ApplyRequirements = p.ApplyRequirements

	v.Name = p.Name

	return v
}
