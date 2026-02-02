# Project Documentation

This directory contains implementation plans and institutional knowledge.

## Structure

- **issues/** - Implementation plans for GitHub issues
- **adhoc/** - Ad-hoc implementation plans
- **knowledge/** - Institutional knowledge and learnings

## Creating Plans

Use `/plan <description or issue-number>` to create new implementation plans.

The planner will automatically create the appropriate directory structure:
- Issue-based: `.docs/issues/<issue-number>/`
- Ad-hoc: `.docs/adhoc/<plan-name>/`

## Documentation Guidelines

### Implementation Plans
- **Purpose**: Document approach for specific feature/issue
- **Lifecycle**: Temporary, archived after implementation
- **Location**: `issues/` or `adhoc/` subdirectories
- **Files**:
  - `<name>-plan.md` - Implementation phases with code
  - `<name>-tasks.md` - Task breakdown
  - `<name>-context.md` - Background and decisions
  - `<name>-research.md` - Research findings

### Knowledge Documentation
- **Purpose**: Permanent, project-wide patterns and learnings
- **Lifecycle**: Living documentation, updated continuously
- **Location**: `knowledge/` subdirectories
- **Categories**:
  - `architecture/` - System design and patterns
  - `conventions/` - Project-specific conventions
  - `learnings/` - Discoveries and corrections
  - `gotchas/` - Known issues and workarounds

See `knowledge/README.md` for detailed knowledge documentation guidelines.

## Workflow

1. Use `/plan` to create implementation plans
2. Document learnings in `knowledge/learnings/` during work
3. Promote learnings to `architecture/` when formalized
4. Track recurring issues in `gotchas/`

See the `iw-workflow` skill (auto-loaded) for complete process documentation.
