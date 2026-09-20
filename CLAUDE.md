# Statusengine Web Interface - Implementation Brief

## 1. Objective
Build a new Statusengine Web Interface that replaces the legacy UI with a modern, responsive, and user-friendly application.

Primary goal:
- Ship a production-ready frontend and backend that read monitoring data from MySQL and expose a clean UI for operations workflows.

## 2. Hard Constraints
- Do not use PrimeNG.
- Do not use Optimus UI.
- Use MySQL as the primary data source.
- Use the `/command` API endpoint for external commands.
- Assume Docker is available for local development and deployment workflows.
- Assume potential access credentials may be stored in `.claude/specs/ressouces.txt`.
- Build the API backend as part of this project (it does not exist yet).

## 2.1 Existing Infrastructure (Given)
The following components already exist and should be treated as upstream systems:
- Naemon
- Statusengine worker
- MySQL database

Only the web interface and its API backend must be implemented in this project.

## 3. Preferred Stack (Soft Preference)
I am most familiar with:
- Angular
- Vue.js
- Go
- CakePHP
- AngularJS

This is not a strict limitation. Choose the stack that best supports long-term maintainability and developer velocity.

## 4. Legacy and Domain References
- Legacy interface reference: https://github.com/statusengine/interface
- Naemon external command reference: https://www.naemon.io/documentation/developer/externalcommands/

Statusengine is a backend for Nagios/Naemon. The new UI must support common operator tasks and monitoring workflows.

## 5. Functional Requirements

### 5.1 Authentication and Authorization
- Implement login-based authentication.
- Implement authorization (role- or permission-based access).
- Provide Guest/Demo mode

### 5.2 Core Pages
- Dashboard
- Hosts list
- Services list
- Host detail view
- Service detail view
- Downtimes list
- Acknowledgements list
- Current problems list
- Log entries list

### 5.3 History Views
Provide history pages for:
- Executed checks
- State changes
- Sent notifications
- Acknowledgements

### 5.4 Metrics / Performance Data
- Display performance data as charts/graphs.
- Use MySQL as current source for metrics.
- Design a provider abstraction so Graphite can be added later with minimal UI changes.

### 5.5 External Commands
Support operator actions through `/command`, including:
- Reschedule host checks
- Reschedule service checks
- Submit passive check results
- Send custom notifications
- Schedule downtimes
- Acknowledge host or service issues

Map supported actions to Naemon command semantics from the official reference.

### 5.6 Optional Real-Time Updates
- Optional: integrate WebSocket API for live status updates.
- If implemented, provide fallback polling when socket is unavailable.

## 6. Non-Functional Requirements
- Fully responsive on desktop, tablet, and mobile.
- Modern UI component system and consistent design language.
- Good performance for large host/service datasets (pagination/filtering/sorting).
- Clear error handling and user feedback.
- Accessibility basics (keyboard navigation, labels, contrast).

## 7. Architecture Guidance
- Separate data access from UI rendering.
- Implement service/repository layer for each domain area (hosts, services, events, metrics, commands).
- Keep metrics source behind an interface to allow MySQL now and Graphite later.
- Keep command execution module isolated and auditable.

## 8. Delivery Expectations
Implement in phases:

### Phase 1 - Foundation
- Project scaffolding
- Auth + Guest/Demo mode
- Base layout, navigation, responsive shell

### Phase 2 - Monitoring Views
- Dashboard
- Hosts/services list and details
- Current problems, downtimes, acknowledgements, log entries

### Phase 3 - History + Metrics
- History pages
- Performance charts from MySQL
- Metrics provider abstraction ready for Graphite adapter

### Phase 4 - Commanding + Realtime
- External command UI and backend integration via `/command`
- Optional WebSocket live updates with fallback

### Phase 5 - Hardening
- UX polish
- Error/edge-case handling
- Basic test coverage and documentation

## 9. Acceptance Criteria
The implementation is complete when:
- All required pages and lists are functional.
- Host and service detail pages show relevant status and history context.
- External commands can be submitted successfully through `/command`.
- Metrics charts render from MySQL data.
- Layout is responsive across common viewport sizes.
- Guest/Demo mode works with enforced restrictions.
- Code structure supports adding Graphite with minimal disruption.

## 10. Output Format Requested from the Implementer
When executing this brief, provide:
- Chosen architecture and stack rationale.
- Folder/module structure.
- API/data contract assumptions.
- Incremental implementation plan.
- Completed features mapped back to sections 5 and 9.
- Known gaps and next steps.