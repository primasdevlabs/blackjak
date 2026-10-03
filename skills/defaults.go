package skills

import (
	"embed"
	"io/fs"
	"path"
	"strings"
)

//go:embed defaults/*/SKILL.md
var builtinSkillsFS embed.FS

// BuiltinSkills returns skills shipped inside the agent binary.
// Available on every fresh install; workspace skills with the same name override.
func BuiltinSkills() []Skill {
	var out []Skill
	_ = fs.WalkDir(builtinSkillsFS, "defaults", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.EqualFold(d.Name(), "SKILL.md") {
			return nil
		}
		data, err := builtinSkillsFS.ReadFile(p)
		if err != nil {
			return nil
		}
		// defaults/<name>/SKILL.md
		name := path.Base(path.Dir(p))
		if name == "" || name == "." || name == "defaults" {
			return nil
		}
		body := string(data)
		out = append(out, Skill{
			Name:        name,
			Description: firstNonEmptyLine(body),
			Body:        body,
			Enabled:     true,
			Path:        "builtin://" + name,
		})
		return nil
	})
	return out
}
