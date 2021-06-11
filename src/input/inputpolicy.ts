import dispatchAgent from './dispatchagent';
import * as S5Err from '../error';
import * as S5Event from '../event';

export default interface inputPolicy {
  AddDispatchAgent(d: dispatchAgent, atFront?: boolean): void;
  RemoveDispatchAgent(d: dispatchAgent): void;
  DispatchAgents(): dispatchAgent[];
  ProcessEvent(e: S5Event.Event): boolean;
}

// this is "default" in the sense of a default implementation
// in the base class, not default with respect to input processing
// policies
export abstract class inputPolicyBase implements inputPolicy {
  agent: dispatchAgent[] = [];

  AddDispatchAgent(d: dispatchAgent, atFront?: boolean): void {
    if (atFront) {
      this.agent.unshift(d);
    } else {
      this.agent.push(d);
    }
  }

  RemoveDispatchAgent(d: dispatchAgent): void {
    const index = this.agent.indexOf(d);
    if (index < 0) {
      throw S5Err.NewError(
        S5Err.Messages.NoAgent,
        'cannot find agent in RemoveDispatchAgent'
      );
    }
    if (this.agent.length === 0) {
      throw S5Err.NewError(
        S5Err.Messages.NoAgent,
        'no policies found in RemoveDispatchAgent'
      );
    }
    this.agent.slice(index, 1);
  }

  // returns the actual list, not a copy!
  DispatchAgents(): dispatchAgent[] {
    return this.agent;
  }

  abstract ProcessEvent(e: S5Event.Event): boolean;
}
