/**
 * rtFlags is for flags that should only be changed by developers since it
 * requires a recompile to change.
 */
export default interface rtFlags {
  /**
   * If Debug is set to true will cause a lot of logging to the logger that starts
   * with the prefix 'EVDEBUG' or 'SEVEN5' so you can easily 'grep -v' if you want. This
   * debug information is probably only of interest to you if you are changing
   * evalvite or seven5 proper.  The debug value defaults to false.  If you are using this
   * feature, you probably want to set the 'name' of all your attributes when
   * calling their constructors.
   */
  Debug: boolean;
  /**
   * If WarnOnUnboundAttributes is set to true will cause a warning when you modify
   * an attribute and there is no interactor hooked to it. this can help debugging
   * if you have forgotten to call BindModelToInteractor.  This defaults to false.
   * Note that if this is set to true, there often will be spurious warnings
   * generated during program startup because there are points in time when
   * attributes have been created and hooked to each other _before_ the
   * call to bindModelToInteractor.
   */
  WarnOnUnboundAttributes: boolean;
  /**
   * If AbortOnCycle is set to true, evalvite will give up when trying to
   * process updates when it detects a cycle in the dependency graph.  If
   * AbortOnCycle is set to false, evalvite will stop processing at the point
   * where the cycle is detected (breaking the cycle) and attempt to print
   * out some information about the cycle it found.  Almost always, cycles
   * in the dependency graph are a programming error, so this defaults to true.
   */
  AbortOnCycle: boolean;
  /**
   * IDCounter is a shared counter for the generation of IDs by various
   * parts of the code.  It is "public" in the sense that it is ok for
   * user level code to read it, and increment it only.
   */
  IdCounter: number;
}

export class flags implements rtFlags {
  _debug = false;

  _unbound = false;

  _errorOnCycle = true;

  _idCounter = 0;

  get Debug(): boolean {
    return this._debug;
  }

  set Debug(d: boolean) {
    this._debug = d;
  }

  get WarnOnUnboundAttributes(): boolean {
    return this._unbound;
  }

  set WarnOnUnboundAttributes(d: boolean) {
    this._unbound = d;
  }

  get AbortOnCycle(): boolean {
    return this._errorOnCycle;
  }

  set AbortOnCyle(d: boolean) {
    this._errorOnCycle = d;
  }

  get IdCounter(): number {
    return this._idCounter;
  }

  set IdCounter(d: number) {
    this._idCounter = d;
  }
}
