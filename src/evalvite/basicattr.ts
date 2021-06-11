import * as S5Conf from '../conf';
import * as S5Err from '../error';
import attribute from './attribute';
import attrPrivateImpl from './attrprivate';
import './typeutils';

let { IdCounter } = S5Conf.Vars;
const { Debug } = S5Conf.Vars;

export default class basicAttribute<T> extends attrPrivateImpl<T> {
  isComputed = false;

  // undefined means "we don't have a value yet"
  cached: T | undefined = undefined;

  // undefined means "we don't have a function yet"
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  _func: ((...a2: any[]) => T) | undefined = undefined;

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  get Func(): (...a2: any[]) => T {
    if (this._func === undefined) {
      throw S5Err.NewError(S5Err.Messages.NoFunc, `${this.DebugName()}.Func`);
    }
    return this._func;
  }

  /**
   * Setting the function for an attribute has five restrictions. One,
   * the function cannot have optional or rest arguments; the number of
   * arguments must be fixed. Second, the number of inputs provided
   * must match the number of function parameters.  This arity is not checked
   * by this function, but is checked by SetInputsAndFunc. Third, all the
   * inputs must be attributes. If you need other kinds of arguments
   * use references to the enclosing object.  Fourth, a parameter
   * foo of type bar to the function (see #1) implies that an input
   * that is of type attribute<bar>.  (The attribute foo's value will be
   * demanded by the parameter passing to this function.) This constraint is not
   * checked by the code. Finally, the function must be side-effect free.
   * @param f
   */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  set Func(f: (...a2: any[]) => T) {
    if (!this.isComputed) {
      this.isComputed = true; // it is now
    } else if (this._func !== undefined && f.length !== this.inEdges.length) {
      throw S5Err.NewError(
        S5Err.Messages.MismatchedParameters,
        `${this.inEdges.length} inputs versus ${f.length} parameters`
      );
    }
    this._func = f;
    this.markDirty();
  }

  get Inputs(): Array<attribute<unknown>> {
    return this.inEdges;
  }

  /**
   * Set the input objects for a a computed attribute. If this is not
   * a computed attribute, this call will throw an error.  Parameters
   * provided must be the same in number as the previous inputs.
   * @param i
   */
  set Inputs(i: Array<attribute<unknown>>) {
    if (this.isComputed) {
      // resetting the inputs on a computed value is tricky
      if (this._func !== undefined && this.inEdges.length !== i.length) {
        throw S5Err.NewError(
          S5Err.Messages.MismatchedParameters,
          `changing number of inputs from ${this.inEdges.length} to ${i.length}`
        );
      }
      if (this._func !== undefined) {
        S5Conf.Warn(
          `changing the inputs to a computed attribute ${this.DebugName()} and this is not typechecked`
        );
      }
      // clear out old inputs
      this.clearAllInEdges(); // implies reciprocal
    }
    // new inputs
    i.forEach((a) => {
      // fixme(iansmith) cast required
      (a as attrPrivateImpl<unknown>).addOutgoing(this);
      this.addIncoming(a as attrPrivateImpl<unknown>);
    });
  }

  SetInputsAndFunc(
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    inputs: Array<attribute<any>>,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    fn: (...a2: any[]) => T
  ): void {
    if (inputs.length !== fn.length) {
      throw S5Err.NewError(
        S5Err.Messages.MismatchedParameters,
        `${this.DebugName()}: number of parameters of function and number of inputs are not the same`
      );
    }
    this.Inputs = inputs;
    this.Func = fn;
  }

  clearAllInEdges(): void {
    this.inEdges.forEach((e) => {
      e.removeOutgoing(this);
      this.removeIncoming(e);
    });
  }

  /**
   * Set simple value converts a computed value to a simple value and
   * assigns v to that simple value.
   */
  SetSimpleValue(v: T): void {
    if (!this.isComputed) {
      S5Conf.Warn(
        `cannot change ${this.DebugName()} to simple value, it already is one; just setting value`
      );
      this.Set(v);
      return;
    }
    this.clearAllInEdges();
    this.isComputed = false;
    this.cached = v;
  }

  /**
   * Constructs a new attribute and makes it a simple value with
   * no set value.  You must set the value to use this attribute as
   * a simple value, or you may set the function it is computed from
   * via SetInputsAndFunc().
   */
  constructor(debugName?: string) {
    super(debugName || `[basic attribute ${IdCounter}]`);
    IdCounter += 1;
    this.isComputed = false;
    this.dirty = false;
  }

  Set(newValue: T): void {
    if (newValue === this.cached) {
      if (Debug) {
        S5Conf.Log(`EVDEBUG: ${this.DebugName()}: ignoring set to same value`);
      }
      return;
    }
    if (Debug) {
      S5Conf.Log(`EVDEBUG: ${this.DebugName()}: set to ${newValue}`);
    }
    const old: T | undefined | null = this.cached;
    this.cached = newValue;
    this.cleanAndNotify();
    if (old === undefined) {
      this.update(this.cached);
    } else {
      this.update(this.cached, old);
    }
  }

  Get(): T {
    if (this.isComputed) {
      return this.getFunctionValue();
    }
    if (Debug) {
      S5Conf.Log(
        `EVDEBUG: ${this.DebugName()}: get of basic attribute: ${this.cached}`
      );
    }
    if (this.cached === undefined) {
      throw S5Err.NewError(
        S5Err.Messages.NullAttr,
        `${this.DebugName} Get of basic attribute whose value is not yet set`
      );
    }
    this.dirty = false; // shouldn't be necessary but...
    return this.cached;
  }

  getFunctionValue(): T {
    if (!this.dirty) {
      if (this.cached === undefined) {
        throw new Error(
          `${this.DebugName()}: basic attribute is not dirty but has no previously cached value`
        );
      }
      if (Debug) {
        S5Conf.Log(
          `EVDEBUG: ${this.DebugName()}: no reason to recompute function, have cached value`
        );
      }
      return this.cached;
    }

    // need to recompute
    if (Debug) {
      S5Conf.Log(
        `EVDEBUG: ${this.DebugName()}: evaluating function parameters`
      );
    }
    // we have disable eslint here because we need to use any to get the typing right
    // on the line below where we assign this params value to actuals
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const params: unknown[] = this.inEdges.map(
      (dep: attribute<unknown>): unknown => {
        const v = dep.Get();
        return v;
      }
    );
    const { _func } = this;

    const actuals: unknown[] = params;
    if (Debug) {
      S5Conf.Log(
        `EVDEBUG: ${this.DebugName()}: evaluating function and returning value`
      );
    }
    if (_func === undefined) {
      throw S5Err.NewError(
        S5Err.Messages.NoFunc,
        `${this.DebugName()}: function not provided`
      );
    }
    const oldValue = this.cached;
    this.cached = _func.apply(this, actuals);
    this.cleanAndNotify();
    if (oldValue === undefined) {
      this.update(this.cached);
    } else {
      this.update(this.cached, oldValue);
    }
    return this.cached;
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  decode(): any {
    return this.Get();
  }

  // eslint-disable-next-line class-methods-use-this
  AttributeName(): string {
    return `BasicAttribute(${this.isComputed ? 'computed' : 'value'})`;
  }

  // eslint-disable-next-line class-methods-use-this
  AllowsSet(): boolean {
    return !this.isComputed;
  }

  // eslint-disable-next-line class-methods-use-this
  typename(): string {
    return `basic attribute`;
  }
}
