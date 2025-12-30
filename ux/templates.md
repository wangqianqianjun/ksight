# Templates Module UX

## Purpose
Manage YAML templates with parameters, versioning, and sharing.

## Layout
- Left sidebar: file tree with pinned templates, my templates, shared templates.
- Main area: metadata and parameter panel on top, YAML editor below.

## Primary Interactions
- Select a template to view metadata and edit YAML.
- Parameter fields appear above the editor.
- Actions: share, save, duplicate, delete.

## Editor Behavior
- YAML editor supports syntax highlighting and validation.
- Parameters use the ~{} notation.
- A preview panel can show resolved values before apply.

## Empty and Error States
- No templates: show "Create your first template" with a primary action.
- Load error: show error banner with retry.

## Success Criteria
- Users can edit and save a template within one minute.
- Parameter use is clear without reading documentation.
