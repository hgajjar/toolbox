package recipe

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type recipe struct {
	env            map[string]string
	commandTimeout int
	sections       []section
}

type section struct {
	name     string
	commands commands
	async    bool
}

type commands struct {
	parallel bool
	commands []command
}

func newRecipe(recipeFile string) *recipe {
	var recipeYaml = yaml.MapSlice{}

	recipe, err := os.ReadFile(recipeFile)
	if err != nil {
		panic(fmt.Errorf("could not find recipe file: %s", recipeFile))
	}
	if err := yaml.Unmarshal(recipe, &recipeYaml); err != nil {
		panic(fmt.Errorf("unable to read installer recipe file: %s", recipeFile))
	}

	return parseRecipeYaml(recipeYaml)
}

func parseRecipeYaml(recipeYaml yaml.MapSlice) *recipe {
	var recipe recipe
	for _, entity := range recipeYaml {

		switch entity.Key {

		case "env":
			for _, e := range entity.Value.(yaml.MapSlice) {
				recipe.env = make(map[string]string)
				recipe.env[e.Key.(string)] = e.Value.(string)
			}

		case "command-timeout":
			recipe.commandTimeout = entity.Value.(int)

		case "sections":
			for _, e := range entity.Value.(yaml.MapSlice) {
				var section section
				section.name = e.Key.(string)
				section.commands = commands{}

				for _, ee := range e.Value.(yaml.MapSlice) {

					switch ee.Key {
					case "async":
						section.async = ee.Value.(bool)
					case "commands":
						for _, c := range ee.Value.(yaml.MapSlice) {
							name := c.Key.(string)

							if name == "parallel" {
								section.commands.parallel = true
								continue
							}

							var cmdStr string
							var depends []string

							for _, cc := range c.Value.(yaml.MapSlice) {
								switch cc.Key {
								case "command":
									cmdStr = cc.Value.(string)
								case "groups":
								case "depends":
									for _, d := range cc.Value.([]any) {
										depends = append(depends, d.(string))
									}
								}
							}

							command := newCommand(name, cmdStr, depends)

							section.commands.commands = append(section.commands.commands, command)
						}
					}
				}

				recipe.sections = append(recipe.sections, section)
			}
		}
	}

	return &recipe
}
