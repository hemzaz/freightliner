# Cleanup Command

You are tasked with performing comprehensive codebase cleanup to remove dead code, unused files, redundant configurations, and build artifacts. Follow the systematic approach defined in the cleanup specifications.

## Command Usage

Available subcommands:
- `/cleanup scan` - Scan and analyze codebase for cleanup opportunities
- `/cleanup execute` - Execute cleanup operations with safety verification
- `/cleanup plan` - Show cleanup plan without executing
- `/cleanup status` - Show current cleanup status and recommendations

## Cleanup Methodology

### Phase 1: High Confidence Removals (Safe)
Execute these automatically with verification:

1. **Build Artifacts**
   - Remove committed binaries from bin/ directory
   - Remove compiled objects (*.exe, *.so, *.dylib)
   - Remove temporary build files
   - Verify regeneration with `make build`

2. **Empty Directories**
   - Remove directories with no files
   - Remove directories that served no active purpose
   - Verify project structure remains functional

3. **Disabled/Superseded Configuration**
   - Remove configuration files for disabled tools
   - Remove redundant configuration files
   - Remove superseded linting configurations
   - Update references to use canonical configurations

4. **Dead Scripts**
   - Remove scripts that are commented out in Makefiles
   - Remove scripts for disabled tools (e.g., staticcheck when replaced)
   - Verify no hidden dependencies exist

### Phase 2: Medium Confidence Removals (Review Required)
Present these for confirmation before removal:

1. **Historical Documentation**
   - Analysis documents that served their purpose
   - Version-specific documentation that's outdated
   - Implementation notes that are no longer relevant

2. **Test Fixtures**
   - Unused test data files
   - Empty test directories
   - Superseded test configurations

3. **Legacy Examples**
   - Outdated example configurations
   - Examples for deprecated features
   - Redundant example files

### Phase 3: Documentation Maintenance
Always execute these updates:

1. **Reference Updates**
   - Update README.md references to removed files
   - Update .claude/structure.md to reflect current state
   - Fix broken links in all documentation

2. **Configuration Updates**
   - Update workflow files to remove references to deleted files
   - Update Makefile comments about removed scripts
   - Update project structure documentation

## Safety Protocols

### Before Each Phase
- Create git checkpoint: `git add . && git commit -m "Checkpoint before cleanup phase X"`
- Verify clean working directory
- Note current branch and status

### After Each Phase
- Run verification commands:
  - `make build` - Verify build still works
  - `make test-ci` - Verify tests still pass
  - `go mod tidy` - Verify dependencies are clean
- If verification fails:
  - Immediately rollback: `git reset --hard HEAD~1`
  - Report the issue and stop cleanup
  - Investigate dependency before proceeding

### Final Verification
- Run complete CI equivalent locally
- Verify all make targets still function
- Check that documentation references are valid
- Confirm no functionality has been lost

## Analysis Framework

### Confidence Level Assignment

**High Confidence (Auto-remove)**:
- Empty directories with no files for 7+ days
- Build artifacts in version control (bin/, *.exe, etc.)
- Configuration files for explicitly disabled tools
- Scripts commented out in Makefiles with replacement noted

**Medium Confidence (Review required)**:
- Documentation files with "ANALYSIS" or "OVERHAUL" in name
- Historical documentation that served a specific purpose
- Test directories that appear unused but might have utility

**Low Confidence (Flag only)**:
- Files referenced in documentation but appear unused
- Configuration files that might be examples for users
- Scripts that aren't clearly active or inactive

### File Category Analysis

**Source Code Files**: 
- Analyze import statements and dependencies
- Check for unreferenced functions or packages
- Look for TODO comments indicating incomplete features

**Configuration Files**:
- Compare multiple config files for the same tool
- Identify superseded configurations
- Check for consistency across environments

**Documentation Files**:
- Cross-reference with current features
- Identify outdated version-specific docs
- Check for broken internal links

**Scripts and Build Files**:
- Parse Makefile for active vs commented targets
- Check script execution permissions and recent usage
- Identify scripts replaced by other tools

## Implementation Steps

When executing cleanup:

1. **Initialize**
   - Load cleanup configuration from .claude/cleanup-config.yml (if exists)
   - Set confidence threshold (default: high)
   - Initialize safety verification commands

2. **Scan Phase**
   - Traverse entire codebase excluding .git, vendor, node_modules
   - Categorize all files by type and usage patterns
   - Build reference map of file dependencies
   - Assign confidence levels based on analysis rules

3. **Plan Phase**
   - Group findings by cleanup phase and confidence level
   - Create execution plan with verification steps
   - Calculate impact analysis for each proposed removal
   - Generate detailed report of proposed changes

4. **Execute Phase**
   - Execute each cleanup phase in order
   - Create git checkpoint before each phase
   - Perform removals for current phase
   - Run verification commands
   - Rollback if verification fails
   - Continue to next phase if verification passes

5. **Report Phase**
   - Generate comprehensive cleanup report
   - Document all changes made
   - Report verification results
   - Provide summary of cleanup impact

## Success Criteria

A cleanup operation is successful when:
- All targeted files/directories are removed appropriately
- All functionality verification passes (build, test, lint)
- All documentation references are updated correctly
- Git history shows clean, logical commits
- No breaking changes are introduced
- Codebase is measurably cleaner and more focused

## Error Handling

If any verification fails:
- Immediately halt cleanup process
- Rollback the failed phase using git reset
- Report the specific failure and its context
- Provide guidance for manual investigation
- Preserve all analysis data for debugging

Remember: Safety is paramount. It's better to leave questionable files than to break functionality. When in doubt, flag for manual review rather than auto-remove.