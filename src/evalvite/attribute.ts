/**
 * An updatable is something that has a UpdateValue() method. If it does, it can
 * connected to evalvite. (Also requires a connect and disconnect, but this
 * can just have the null implementation.) This helps break a number of import cycles that
 * would be created if we used interactors in defining where evalvite output
 * goes.
 *
 * Note that the two value parameters to UpdateValue() are optional.  On a normal
 * Set() of the value of source both new and old are provided. If it is
 * the first Set() of source, we have no old value and pass just new value.
 * If neither are passed the receiver is being notified that the value of
 * source is out of date, and you may want to request it (eagerness).
 */
export interface updateable<T> {
  ConnectAttribute(source:attribute<T>):void;
  DisconnectAttribute(source:attribute<T>):void;
  UpdateValue(source:attribute<T>, newValue?: T, oldValue?:T):void;
}

// Attribute is a wrapper around a value.  Using get() and set()
export default interface attribute<T> {
  // debugName will be displayed in log messages when you turn on debug mode.
  DebugName(): string;
  Get(): T;
  Set(v: T): void;
  // addInteractor connects an interactor to this attribute.  When the attribute
  // is dirty, the component's visuals are dirty and forceUpdate() is used to
  // cause a redraw.
  AddUpdateable(u: updateable<T>): void;
  RemoveUpdateable(u: updateable<T>): void;

  // Eager attributes do not exploit laziness and immediately update
  // once they realize that one of their dependencies is OOD.  If you
  // are doing things that are visible to the user, you should use Eager.
  // Users don't want their display not updating until a value gets used.
  Eager:boolean;

  SetInputsAndFunc(
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    inputs: Array<attribute<any>>,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    fn: (...a2: any[]) => T
  ): void;
  // Drop connection looks for connections from *this* to each of
  // of the provided attributes and if it finds any, it removes.
  // If no connection is found, it ignores.
  DropAttributeConnection(...other:attribute<unknown>[]):void;
  /**
   * Returns true if this attribute's value can be set.  Normally, this
   * means that this attribute is a simple (source) attribute.
   */
  AllowsSet():boolean;
}
