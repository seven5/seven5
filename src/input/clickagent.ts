import dispatchAgent, { dispatchAgentBase } from './dispatchagent';
import * as S5Event from '../event';
import * as S5Interactor from '../interactor';
import * as S5Conf from '../conf';
import * as S5Err from '../error';

/**
 * Protocol for Clickable interactors.  These methods (this type)
 * should be implemented if you want the input dispatch mechanism
 * to make you eligible for clicks.  Only start and end are required.
 */
// xxx fixme(iansmith): should this require that it extends Interactor?
export interface Clickable extends S5Interactor.Interactor {
  ClickStart(e: S5Event.Event): void;
  ClickEnd(e: S5Event.Event): void;

  // these three are optional
  ClickDrag(e: S5Event.Event): void;
  ClickEnter(e: S5Event.Event): void;
  ClickLeave(e: S5Event.Event): void;
}

// only start and end are required
const isClickable = (i: S5Interactor.Interactor): i is Clickable => {
  return 'ClickStart' in i && 'ClickEnd' in i;
};

export default class clickAgent
  extends dispatchAgentBase
  implements dispatchAgent {
  // this is non null when we have received the mouseDown but not
  // yet the mouseUp
  _focus: Clickable | null = null;

  get focus(): Clickable | null {
    return this._focus;
  }

  set focus(c: Clickable | null) {
    this._focus = c;
  }

  Dispatch(e: S5Event.Event): boolean {
    if (e.PickList === null) {
      return false; // nothing to do unless we are getting this via position policy
    }
    if (this.focus === null && e.Type === S5Event.Name.MouseButtonDown) {
      return this.dispatchFirst(e);
    }
    if (this.focus !== null && e.Type === S5Event.Name.MouseButtonDown) {
      S5Conf.Warn(
        'ClickAgent received MouseDown when focus is already established,ignoring'
      );
      return true;
    }
    if (
      e.Type === S5Event.Name.MouseButtonUp ||
      e.Type === S5Event.Name.MouseMove
    ) {
      if (this.focus === null) {
        // you could make the case we should return true here because
        // and therefore stop the processing of these moves or ups
        return false;
      }
      return this.dispatchNotFirst(e);
    }
    return false;
  }

  dispatchFirst = (e: S5Event.Event): boolean => {
    if (e.PickList === null) {
      // this is here to make the compiler happy
      throw S5Err.NewError(
        S5Err.Messages.BadState,
        "can't happen in dispatchFirst"
      );
    }
    const first = e.PickList.FindFirst((p: S5Event.Pickable): boolean => {
      // xxx fixme(iansmith) cast... can it be avoided? why do I need the ?, shouldn't it work like above?
      return isClickable(p as S5Interactor.Interactor);
    });
    // is anything clickable in that pick list?
    if (first === null) {
      return false; // nope
    }
    // xxx fixme(iansmith) because of cast... can it be avoided?
    const frontClickable = first as Clickable;
    // we checked for bad state machine state above in Dispatch()
    const clickableFocus: Clickable | null = (frontClickable as unknown) as Clickable;
    this.focus = clickableFocus;
    frontClickable.ClickStart(e);
    return true;
  };

  // events other than the first mouseDown
  dispatchNotFirst(e: S5Event.Event): boolean {
    if (this.focus === null) {
      // should never happen, see Dispatch()
      S5Conf.Warn(`Bad state machine state! No focus for event ${e.Type}`);
      return false;
    }
    switch (e.Type) {
      case S5Event.Name.MouseMove:
        if ('ClickDrag' in this.focus) {
          this.focus.ClickDrag(e);
          return true;
        }
        return false;
      case S5Event.Name.MouseButtonUp:
        if ('ClickEnd' in this.focus) {
          this.focus.ClickEnd(e);
          this.focus = null;
          return true;
        }
        // the type system should really prevent this ever happening
        S5Conf.Warn(
          `Bad clickable definition! Can't find ClickEnd on ${this.focus}`
        );
        this.focus = null;
        return true;
      default:
        return false;
    }
  }
}
