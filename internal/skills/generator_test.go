package skills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gleanwork/glean-cli/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Register minimal test schemas so these tests don't depend on cmd/ init().
// init() in a _test.go file only runs when compiling tests in this package.
func init() {
	schema.Register(schema.CommandSchema{
		Command:     "testsearch",
		Description: "Test search description.",
		WhenToUse:   "Find test things.",
		Surface:     schema.SurfacePlatform,
		Flags: map[string]schema.FlagSchema{
			"--query": {Type: "string", Description: "Search query", Required: true},
		},
		Example: "glean testsearch --query hello",
	})
	schema.Register(schema.CommandSchema{
		Command:     "testpins",
		Description: "Test pins description.",
		Surface:     schema.SurfaceLegacy,
		Flags:       map[string]schema.FlagSchema{},
		Example:     "glean testpins list",
	})
}

func runGenerator(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, Generate(dir))
	return dir
}

func readRootSkill(t *testing.T, dir string) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(dir, rootSkillName, "SKILL.md"))
	require.NoError(t, err)
	return string(body)
}

func TestGenerate_WritesRootSkill(t *testing.T) {
	dir := runGenerator(t)

	p := filepath.Join(dir, rootSkillName, "SKILL.md")
	info, err := os.Stat(p)
	require.NoError(t, err, "root SKILL.md must exist")
	assert.Greater(t, info.Size(), int64(0), "root SKILL.md must be non-empty")
}

func TestGenerate_PointsToAgentHelpAsSourceOfTruth(t *testing.T) {
	content := readRootSkill(t, runGenerator(t))

	assert.Contains(t, content, "glean agent-help")
	assert.Contains(t, strings.ToLower(content), "source of truth")
	assert.Contains(t, content, "--json")
}

func TestGenerate_CommandMapFromRegistry(t *testing.T) {
	content := readRootSkill(t, runGenerator(t))

	assert.Contains(t, content, "`glean testsearch`")
	assert.Contains(t, content, "Find test things.")
	assert.Contains(t, content, schema.SurfacePlatform)
	// testpins has no WhenToUse — Description is the fallback.
	assert.Contains(t, content, "Test pins description.")
}

func TestGenerate_NoReferenceFiles(t *testing.T) {
	dir := runGenerator(t)

	_, err := os.Stat(filepath.Join(dir, rootSkillName, "reference"))
	assert.True(t, os.IsNotExist(err), "per-command reference files are retired in favor of agent-help")
}

func TestGenerate_RemovesRetiredReferenceDir(t *testing.T) {
	dir := t.TempDir()
	stale := filepath.Join(dir, rootSkillName, "reference")
	require.NoError(t, os.MkdirAll(stale, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(stale, "search.md"), []byte("stale"), 0o644))

	require.NoError(t, Generate(dir))

	_, err := os.Stat(stale)
	assert.True(t, os.IsNotExist(err), "retired reference dir must be cleaned up")
}

func TestGenerate_NoLegacySkillDirs(t *testing.T) {
	dir := runGenerator(t)

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if e.Name() == rootSkillName {
			continue
		}
		assert.NotContains(t, e.Name(), skillPrefix,
			"unexpected legacy skill dir after generation: %s", e.Name())
	}
}

func TestGenerate_Idempotent(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, Generate(dir))
	first := snapshotTree(t, dir)

	require.NoError(t, Generate(dir))
	second := snapshotTree(t, dir)

	assert.Equal(t, first, second, "Generate must be idempotent")
}

func TestCleanStaleSkillDirs_RemovesLegacy(t *testing.T) {
	dir := t.TempDir()
	for _, cmd := range []string{"search", "chat"} {
		stale := filepath.Join(dir, skillPrefix+cmd)
		require.NoError(t, os.MkdirAll(stale, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(stale, "SKILL.md"), []byte("stale"), 0o644))
	}

	require.NoError(t, cleanStaleSkillDirs(dir))

	for _, cmd := range []string{"search", "chat"} {
		stale := filepath.Join(dir, skillPrefix+cmd)
		_, err := os.Stat(stale)
		assert.True(t, os.IsNotExist(err), "legacy dir %s should be removed", stale)
	}
}

func TestCleanStaleSkillDirs_PreservesRoot(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, rootSkillName)
	require.NoError(t, os.MkdirAll(root, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "SKILL.md"), []byte("# root"), 0o644))

	require.NoError(t, cleanStaleSkillDirs(dir))

	info, err := os.Stat(root)
	require.NoError(t, err, "root dir must survive cleanup")
	assert.True(t, info.IsDir())
}

func TestCleanStaleSkillDirs_PreservesUnrelatedDirs(t *testing.T) {
	dir := t.TempDir()
	unrelated := filepath.Join(dir, "some-users-custom-dir")
	require.NoError(t, os.MkdirAll(unrelated, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(unrelated, "file.md"), []byte("keep me"), 0o644))

	require.NoError(t, cleanStaleSkillDirs(dir))

	_, err := os.Stat(unrelated)
	assert.NoError(t, err, "unrelated dir must not be touched by cleanup")
}

func TestRootSkill_IncludesMigrationNote(t *testing.T) {
	content := readRootSkill(t, runGenerator(t))
	assert.Contains(t, content, "npx -y skills remove",
		"root SKILL.md must carry the per-command cleanup one-liner")
}

// snapshotTree returns a map of relative-path -> file contents for every file
// under root. Used by idempotency tests.
func snapshotTree(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out[rel] = string(body)
		return nil
	})
	require.NoError(t, err)
	return out
}

// Verify the retired strings module-wide: the generator must no longer strip
// frontmatter into reference files or reference glean schema.
func TestRootSkill_NoStaleSchemaReferences(t *testing.T) {
	content := readRootSkill(t, runGenerator(t))
	assert.False(t, strings.Contains(content, "glean schema "),
		"skill must route agents to agent-help, not the deprecated schema command")
	assert.NotContains(t, content, "(reference/",
		"skill must not link to retired reference files")
}
