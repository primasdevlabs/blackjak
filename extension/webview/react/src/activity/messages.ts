import { ActivityCategory } from './activityTypes';

export const ACTIVITY_MESSAGES: Record<ActivityCategory, string[]> = {
  orchestrating: [
    "Orchestrating the next move…",
    "Planning the route through the code…",
    "Putting the pieces in order…",
    "Setting up the next phase…",
    "Balancing execution steps…",
  ],

  exploring: [
    "Stargazing at the codebase…",
    "Following the trail…",
    "Mapping the territory…",
    "Looking under the hood…",
    "Scanning project files…",
    "Examining module boundaries…",
  ],

  thinking: [
    "Thinking through the dependencies…",
    "Connecting a few loose ends…",
    "Reading between the lines…",
    "Synthesizing context…",
    "Tracing logical pathways…",
  ],

  debugging: [
    "Tracing the rabbit hole…",
    "Narrowing things down…",
    "Hunting for the mismatch…",
    "Untangling this one…",
    "Cross-checking the evidence…",
    "Circling back to the failing test…",
  ],

  testing: [
    "Putting the hypothesis to the test…",
    "Checking what changed…",
    "Cross-checking test results…",
    "Evaluating test assertions…",
    "Running validation suite…",
  ],

  editing: [
    "Making the surgical changes…",
    "Putting the fix in place…",
    "Reshaping the troublesome bits…",
    "Updating codebase files…",
    "Refactoring workspace modules…",
  ],

  finishing: [
    "Taking another pass…",
    "Giving everything a final look…",
    "Making sure the pieces fit…",
    "Finalizing task execution…",
  ],
};
