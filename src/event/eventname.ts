// builtins are listed here and shoved into the default generator
// inside policyManager.
//
// this is const because we don't want you adding to OUR list of
// event names, because you can add your own in eventGenerator.
const eventName = {
  MouseButtonDown: 'mouseButtonDown',

  MouseMove: 'mouseMove',

  MouseButtonUp: 'mouseButtonUp',

  // N.B. enter/leave refer to the root
  MouseEnter: 'mouseEnter',

  // N.B. enter/leave refer to the root
  MouseLeave: 'mouseLeave',
} as const;

export default eventName;
