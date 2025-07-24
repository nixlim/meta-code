# Meta-MCP Server Development Workflow - July 2025

## Development Environment & Tools

### Core Development Stack
- **Language**: Go 1.24.2
- **Platform**: Darwin (macOS) 24.5.0
- **VCS**: Git with structured commit messages
- **Testing**: Go test framework with race detection
- **Coverage**: Target >80% per package, >85% overall

### Claude Code Integration
The project uses Claude Code as the primary development agent with sophisticated tooling:

#### MCP Tools
1. **Serena MCP**: Code navigation and memory management
   - Project traversal and symbol lookup
   - Memory persistence across sessions
   - Onboarding and context management

2. **Zen MCP**: Advanced development capabilities
   - Debug & root cause analysis
   - Code review workflows
   - Test generation
   - Performance analysis
   - Security auditing

3. **Browser Tools MCP**: UI/UX testing
   - Console log analysis
   - Network monitoring
   - Accessibility audits
   - Performance profiling

#### Command System Structure
Located in `.claude/commands/`, organized by function:

1. **Coordination** (`coordination/`)
   - `swarm-init.md`: Initialize swarm topology
   - `swarm-status.md`: Monitor agent activity
   - `agent-spawn.md`: Create specialized agents

2. **Automation** (`automation/`)
   - `auto-spawn.md`: Automatic agent creation
   - `workflow-select.md`: Choose execution strategy
   - `task-execute.md`: Run automated tasks

3. **Analysis** (`analysis/`)
   - `performance-report.md`: System metrics
   - `bottleneck-detect.md`: Identify issues
   - `token-usage.md`: Track AI usage

4. **GitHub Integration** (`github/`)
   - `repo-analyze.md`: Repository insights
   - `pr-review.md`: Automated reviews
   - `code-quality.md`: Quality metrics

5. **SPARC Modes** (`sparc/`)
   - Development modes for specific tasks
   - API development workflows
   - UI component building

## Development Workflows

### 1. Test-Driven Development (TDD)
```
1. Write failing test
2. Implement minimal code to pass
3. Refactor with confidence
4. Maintain >80% coverage
```

### 2. Swarm-Based Development
For complex tasks, use hierarchical swarm topology:
- **Coordinator**: Orchestrates overall task
- **Coders**: Implement specific components
- **Analysts**: Review and optimize
- **Testers**: Verify functionality
- **Reviewers**: Ensure code quality

### 3. Memory-Driven Development
- **Memory Bank** (`/memory-bank/`): Project context
- **Serena Memories** (`.serena/memories/`): Development history
- **Activity Log** (`.claude-updates`): Change tracking

### 4. Code Review Process
Automated review workflow using Zen MCP:
1. Implement changes
2. Run `zen:codereview` for analysis
3. Address findings
4. Verify with `zen:precommit`
5. Document in `.claude-updates`

## Quality Standards

### Code Organization
1. **Package Structure**: Clear boundaries and interfaces
2. **Error Handling**: Consistent wrapping with context
3. **Naming**: Follow Go idioms and conventions
4. **Documentation**: Comprehensive godoc comments

### Testing Standards
1. **Unit Tests**: Table-driven with comprehensive scenarios
2. **Integration Tests**: Real-world use cases
3. **Benchmarks**: Performance-critical paths
4. **Race Detection**: All tests pass with `-race`

### Error Management
- Standardized error codes (constants)
- Contextual error wrapping
- Proper error types for different scenarios
- Comprehensive error testing

## Recent Workflow Improvements

### Control Structure Refactor (July 2025)
- Reorganized `.claude/commands/` for efficiency
- Enhanced swarm orchestration capabilities
- Improved automation workflows
- Streamlined development patterns

### Testing Framework Enhancement
- Centralized test utilities in `test/testutil/`
- JSON fixtures for test data
- Concurrent testing helpers
- Mock implementations for all interfaces

### Code Quality Automation
- Automated error code standardization
- Duplicate code detection and removal
- Coverage reporting and analysis
- Performance benchmark tracking

## Common Development Patterns

### 1. Feature Implementation
```
1. Update memory bank context
2. Create tests following TDD
3. Implement with swarm if complex
4. Run comprehensive tests
5. Code review with Zen
6. Update documentation
7. Log in .claude-updates
```

### 2. Bug Fixing
```
1. Use zen:debug for root cause
2. Write failing test case
3. Implement fix
4. Verify all tests pass
5. Document fix in updates
6. Update relevant memories
```

### 3. Performance Optimization
```
1. Run benchmarks
2. Use zen:analyze for bottlenecks
3. Implement optimizations
4. Verify improvements
5. Document changes
```

## CI/CD Integration Points

### Pre-commit Checks
- Test coverage verification
- Linting and formatting
- Race condition detection
- Build verification

### Continuous Integration
- Automated test runs
- Coverage reporting
- Performance benchmarks
- Security scanning

### Documentation Updates
- Auto-generate from godoc
- Update memory banks
- Maintain change logs
- Keep README current

## Collaboration Patterns

### Asynchronous Development
- Memory banks provide context
- Detailed commit messages
- Comprehensive documentation
- Activity logging

### Knowledge Transfer
- Serena memories for persistence
- Detailed code comments
- Architecture documentation
- Testing examples

## Troubleshooting Guide

### Common Issues
1. **Build Failures**: Check transport package fix
2. **Test Failures**: Verify race conditions
3. **Coverage Drops**: Add missing test cases
4. **Memory Gaps**: Update memory banks

### Debug Workflow
1. Identify issue with logs
2. Use zen:debug for analysis
3. Isolate with unit tests
4. Fix and verify
5. Document resolution