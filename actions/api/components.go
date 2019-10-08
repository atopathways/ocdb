package api

import (
	"fmt"

	"github.com/gobuffalo/buffalo"
	"github.com/isimluk/ocdb/pkg/masonry"
	"github.com/opencontrol/compliance-masonry/pkg/lib/common"
)

// ComponentsResource show components like defined by opencontrols
type ComponentsResource struct {
	buffalo.Resource
}

// List default implementation.
func (v ComponentsResource) List(c buffalo.Context) error {
	ms := masonry.GetInstance()
	return c.Render(200, r.JSON((*ms).GetAllComponents()))
}

// Show default implementation.
func (v ComponentsResource) Show(c buffalo.Context) error {
	ms := masonry.GetInstance()
	component, found := (*ms).GetComponent(c.Param("component_id"))
	if found {
		return c.Render(200, r.JSON(component))
	}
	return c.Render(404, r.JSON("Not found"))
}

// CustomControl is object that ties together information from standard with product specific "satisfaction" description
type CustomControl struct {
	Key       string
	Control   common.Control
	Satisfies common.Satisfies
}

func standardToLogicalView(s common.Standard) map[string][]CustomControl {
	result := make(map[string][]CustomControl)
	controls := s.GetControls()
	for _, controlName := range s.GetSortedControls() {
		control := controls[controlName]
		_, ok := result[control.GetFamily()]
		if !ok {
			result[control.GetFamily()] = make([]CustomControl, 0)
		}
		result[control.GetFamily()] = append(result[control.GetFamily()], CustomControl{
			Key:     controlName,
			Control: control})
	}
	return result
}

func logicalView(ms *common.Workspace, c common.Component) (map[string]map[string][]CustomControl, []string) {
	result := make(map[string]map[string][]CustomControl)
	problems := make([]string, 0)

	for _, satisfy := range c.GetAllSatisfies() {
		standardKey := satisfy.GetStandardKey()
		_, ok := result[standardKey]
		if !ok {
			standard, found := (*ms).GetStandard(standardKey)
			if found {
				result[standardKey] = standardToLogicalView(standard)
			}

		}
		found := false
		for groupId, group := range result[standardKey] {
			for i, cc := range group {
				if cc.Key == satisfy.GetControlKey() {
					if cc.Satisfies != nil {
						problems = append(problems, fmt.Sprintf("Found duplicate item: %s", cc.Key))
					}

					result[standardKey][groupId][i].Satisfies = satisfy
					found = true
					break
				}

			}
			if found {
				break
			}
		}
		if !found {
			problems = append(problems, fmt.Sprintf("Could not found reference %s in the standard %s", satisfy.GetControlKey(), standardKey))

		}
	}

	return result, problems
}

// ComponentControlsHandler gives logical human readable view of open control items available.
func ComponentControlsHandler(c buffalo.Context) error {
	ms := masonry.GetInstance()
	component, found := (*ms).GetComponent(c.Param("component_id"))
	if found {
		lv, problems := logicalView(ms, component)
		result := make(map[string]interface{})
		result["name"] = component.GetName()
		result["controls"] = lv
		result["errors"] = problems

		return c.Render(200, r.JSON(result))
	}
	return c.Render(404, r.JSON("Not found"))
}