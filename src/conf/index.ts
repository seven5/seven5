// don't try to get too fancy, you can sub in your own logging machinery if you want
// by assigning to these vars.

// we WANT to use the console here because it's logging
/* eslint-disable no-console */

// it's not usually a good idea to have mutable exports but we *intend* this
// to be something that the user can install their own code into
/* eslint-disable import/no-mutable-exports */
/* eslint-disable no-var */

import { flags } from './vars';

/**
 * Log sends a log message to the console by default.  Most messages to the console
 * should use this level.
 * @param msg required first parameter
 * @param extra any number of extra parameters
 */
export var Log = (msg: unknown, ...extra: unknown[]): void => {
  console.log(msg, ...extra);
};
/**
 * Warn sends a warning to the console by default.   This should be used to
 * notify the developer or user that 1) there is a serious problem and 2) there
 * are ways to remedy it.
 * @param msg required first parameter
 * @param extra any number of extra parameters
 */
export var Warn = (msg: unknown, ...extra: unknown[]): void => {
  console.warn(msg, ...extra);
};
/**
 * Error sends an error to the console by default.   This should be used when
 * reporting things that render the software unusable (too broken to continue).
 * @param msg required first parameter
 * @param extra any number of extra parameters
 */
export var Error = (msg: unknown, ...extra: unknown[]): void => {
  console.error(msg, ...extra);
};

export const Vars = new flags();
