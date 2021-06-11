import * as S5Event from '../event';

export default interface dispatchAgent {
  /**
  /**
   * Return true if you handled the event and no more processing should. Return
   * false for processing to continue.
   * @param e
   */
  Dispatch(e: S5Event.Event): boolean;
}

export abstract class dispatchAgentBase {
  abstract Dispatch(e: S5Event.Event): boolean;
}
