package helpers

import (
	"regexp"

	"github.com/manifoldco/promptui"
)

func GetVariant(prompt string, variants interface{}, template string) (int, error) {
	re := regexp.MustCompile(`(?m)({{.*?)(}})`)

	activeTemplate := string(re.ReplaceAll([]byte(template), []byte("${1} | green ${2}")))
	inactiveTemplate := string(re.ReplaceAll([]byte(template), []byte("${1} | white ${2}")))
	v_prompt := promptui.Select{
		Label: prompt,
		Items: variants,
		Templates: &promptui.SelectTemplates{
			Active:   activeTemplate,
			Inactive: inactiveTemplate,
		},
	}
	user_input, _, err := v_prompt.Run()
	return user_input, err
}

func GetString(prompt string, secure bool) (string, error) {
	var s_prompt promptui.Prompt
	if secure {
		s_prompt = promptui.Prompt{
			Label: prompt,
			Mask:  '*',
		}
	} else {
		s_prompt = promptui.Prompt{
			Label: prompt,
		}
	}
	return s_prompt.Run()
}
