import { ActivityCategory } from './activityTypes';

/** Professional status lines shown while the agent is working. */
export const ACTIVITY_MESSAGES: Record<ActivityCategory, string[]> = {
  orchestrating: [
    'Coordinating next steps…',
    'Preparing the execution plan…',
    'Sequencing work items…',
    'Initializing the next phase…',
    'Organizing task dependencies…',
  ],

  exploring: [
    'Reviewing project structure…',
    'Inspecting relevant files…',
    'Mapping module dependencies…',
    'Scanning workspace files…',
    'Examining code boundaries…',
  ],

  thinking: [
    'Analyzing dependencies…',
    'Evaluating design options…',
    'Synthesizing context…',
    'Tracing control flow…',
    'Assessing implementation impact…',
  ],

  debugging: [
    'Isolating the failure…',
    'Narrowing the root cause…',
    'Comparing expected and actual behavior…',
    'Reviewing error evidence…',
    'Checking related test failures…',
  ],

  testing: [
    'Running validation…',
    'Verifying recent changes…',
    'Evaluating test results…',
    'Checking assertions…',
    'Executing the test suite…',
  ],

  editing: [
    'Applying code changes…',
    'Updating project files…',
    'Refactoring affected modules…',
    'Writing implementation updates…',
    'Synchronizing file edits…',
  ],

  finishing: [
    'Performing a final review…',
    'Confirming completion criteria…',
    'Closing out remaining work…',
    'Finalizing the task…',
  ],
};
