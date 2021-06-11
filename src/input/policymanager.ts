import * as S5Err from '../error';
import * as S5Event from '../event';
import * as S5Interactor from '../interactor';

import inputPolicy from './inputpolicy';
import focusPolicy from './focuspolicy';
import positionalPolicy from './positionalpolicy';
import clickAgent from './clickagent';

export default class policyManager {
  _activePolicies = [] as inputPolicy[];

  ActiveInputPolicies = (): inputPolicy[] => {
    return this._activePolicies;
  };

  /**
   * This doesn't check for duplicates.  xxx fixme(iansmith)
   */
  AddActiveInputPolicy = (p: inputPolicy, atFront?: boolean): void => {
    if (atFront) {
      this._activePolicies.unshift(p);
    } else {
      this._activePolicies.push(p);
    }
  };

  /**
   * Aborts with error if policy not found.  You really shouldn't
   * need to use this very often and never when you don't know what
   * is already in the list of policies.
   */
  RemoveActiveInputPolicy = (p: inputPolicy): void => {
    const index = this._activePolicies.indexOf(p);
    if (index < 0) {
      throw S5Err.NewError(
        S5Err.Messages.NoPolicy,
        'cannot find policy in RemoveActiveInputPolicy'
      );
    }
    if (this._activePolicies.length === 0) {
      throw S5Err.NewError(
        S5Err.Messages.NoPolicy,
        'no policies found in RemoveActiveInputPolicy'
      );
    }
    this._activePolicies.slice(index, 1);
  };

  _eventGenerator: S5Event.Generator = new S5Event.Generator();

  /**
   * This should probably only be used in a case where you are
   * adding or removing events from the event generator.  This should
   * be rarely used.
   */
  get EventGenerator(): S5Event.Generator {
    return this._eventGenerator;
  }

  InitializePoliciesAndDispatchAgents = (r: S5Interactor.Root): void => {
    this.EventGenerator.set(
      S5Event.Name.MouseButtonDown,
      new S5Event.ButtonDownDesc()
    );
    this.EventGenerator.set(
      S5Event.Name.MouseButtonUp,
      new S5Event.ButtonUpDesc()
    );
    this.EventGenerator.set(
      S5Event.Name.MouseMove,
      new S5Event.MouseMoveDesc()
    );
    this.EventGenerator.set(
      S5Event.Name.MouseEnter,
      new S5Event.MouseEnterDesc()
    );
    this.EventGenerator.set(
      S5Event.Name.MouseLeave,
      new S5Event.MouseLeaveDesc()
    );

    const focus = new focusPolicy();
    const pos = new positionalPolicy(r);

    this.AddActiveInputPolicy(focus);
    this.AddActiveInputPolicy(pos);

    pos.AddDispatchAgent(new clickAgent());

    r.Element.onmousedown = (ev: MouseEvent): void => {
      this.ProcessEvent(
        this.EventGenerator.Generate(ev, S5Event.Name.MouseButtonDown)
      );
      r.Mainloop();
    };
    r.Element.onmouseup = (ev: MouseEvent): void => {
      this.ProcessEvent(
        this.EventGenerator.Generate(ev, S5Event.Name.MouseButtonUp)
      );
      r.Mainloop();
    };
    r.Element.onmouseenter = (ev: MouseEvent): void => {
      this.ProcessEvent(
        this.EventGenerator.Generate(ev, S5Event.Name.MouseEnter)
      );
      r.Mainloop();
    };
    r.Element.onmouseleave = (ev: MouseEvent): void => {
      this.ProcessEvent(
        this.EventGenerator.Generate(ev, S5Event.Name.MouseLeave)
      );
      r.Mainloop();
    };
    r.Element.onmousemove = (ev: MouseEvent): void => {
      this.ProcessEvent(
        this.EventGenerator.Generate(ev, S5Event.Name.MouseMove)
      );
      r.Mainloop();
    };
  };

  ProcessEvent(e: S5Event.Event): boolean {
    let found = false;

    this._activePolicies.forEach((p: inputPolicy): void => {
      if (!found) {
        found = p.ProcessEvent(e);
      }
    });

    return found;
  }
}

let singletonPM: policyManager | null = null;

export const createInputPolicyManager = (): policyManager => {
  singletonPM = new policyManager();
  return singletonPM;
};

export const inputPolicyManager = (): policyManager => {
  if (singletonPM === null) {
    throw S5Err.NewError(
      S5Err.Messages.BadState,
      'policy manager is null, did you forget to call CreateInputPolicyManager?'
    );
  }
  return singletonPM;
};
