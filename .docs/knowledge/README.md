# Project Knowledge

This directory contains project-specific technical knowledge, architectural decisions, and institutional knowledge.

For general development standards, see `/CLAUDE.md`.

## Directory Structure

### `/architecture/` - Architectural Decisions and Patterns

Formalized system architecture and design patterns.

### `/conventions/` - Project-Specific Conventions

Project-specific naming, structure, and coding conventions that differ from general standards.

### `/learnings/` - Learnings & Corrections

Discoveries and corrections from planning and implementation. See `learnings/README.md` for documentation format.

### `/gotchas/` - Known Issues and Workarounds

Persistent issues, limitations, and their workarounds.

## When to Add to This Directory

Add to `.docs/knowledge/` when you discover:

- **Project-specific patterns** that differ from general practices
- **System limitations**
- **Architectural decisions** that affect multiple features
- **Common gotchas** that future developers should know
- **Integration patterns** specific to this project

**Do NOT add:**
- General development standards → Use `CLAUDE.md`
- Temporary research notes → Use `<name>-research.md` in plan files
- Design decisions for a specific plan → Use `<name>-context.md` in plan files

## Contributing

When you discover new project-specific knowledge during implementation:

1. Check if it belongs in `.docs/knowledge/` (project-specific) vs `CLAUDE.md` (general)
2. Create or update the appropriate document
3. Add entry to this README
4. Keep documents focused and scannable
