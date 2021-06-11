import SimpleAttribute from './simpleattr';
import computedAttribute from './computedattr';
import arrayAttribute from './arrayattr';
import naiveArrayAttribute from './naivearrayattr';
import equalAttribute from './equalattr';
import minMaxAttribute from './minmaxattr';
import attribute, { updateable } from './attribute';
import * as S5Conf from '../conf';
import { decodeTypename, decodeValue } from './decode';
import { bind, unbind } from './typeutils';
import attrPrivate from './base';

// in case you prefer the model terminology
export type evModel = Record<string, undefined>;

/// ////////////////
// User visible flags
/// ///////////////

function setDebug(d: boolean): void {
  S5Conf.Vars.Debug = d;
}

function setWarnOnUnboundAttributes(d: boolean): void {
  S5Conf.Vars.WarnOnUnboundAttributes = d;
}

/// ////////////////
// Convience functions to create various attribute types and connect to interactors
/// ///////////////
function simple<T>(t: T, debug?: string): attribute<T> {
  return new SimpleAttribute<T>(t, debug);
}

function minMax(
  min: number,
  max: number,
  base: attribute<number>,
  debugName?: string
): minMaxAttribute {
  // i am utterly unable to explain why this "as" is necessary... although it
  // is safe as long as all attribute types are derived from attrPrivateImpl as
  // they are now, I don't understand what the compiler is trying to tell me.
  // ps. can't use an interface here because it doesn't exist at run time?
  return new minMaxAttribute(min, max, base as attrPrivate<number>, debugName);
}

function computed<T>(
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  fn: (...restArgs: any[]) => T,
  inputs: attribute<unknown>[],
  name?: string
): attribute<T> {
  return new computedAttribute<T>(fn, inputs, name);
}

// xxx again, why do I have to do this with the as? because attribute<name>
// xxx is erased at run-time?
function equal(a: attribute<number>, name?: string): attribute<number> {
  return new equalAttribute(a as attrPrivate<number>, name);
}

function array<T extends evModel>(debugName?: string): arrayAttribute<T> {
  return new arrayAttribute<T>(debugName);
}

function naivearray<T>(debugName?: string): naiveArrayAttribute<T> {
  return new naiveArrayAttribute<T>(debugName);
}

export {
  /**
   * SetDebug set to true will cause a lot of logging to the logger that starts
   * with the prefix 'EVDEBUG' so you can easily 'grep -v' if you want. This
   * debug information is probably only of interest to you if you are changing
   * evalvite itself.  The debug value defaults to false.  If you are using this
   * feature, you probably want to set the 'name' of all your attributes when
   * calling their constructors.
   */
  setDebug as SetDebug,
  /**
   * SetWarnOnUnboundAttributes set to true will cause a warning when you modify
   * an attribute and there is no component hooked to it. this can help debugging
   * if you have forgotten to call bindModelToComponent.  This defaults to false.
   * Note that if this is set to true, there often will be spurious warnings
   * generated during program startup because there are points in time when
   * attributes have been created and hooked to each other _before_ the
   * call to bindModelToComponent.
   */
  setWarnOnUnboundAttributes as SetWarnOnUnboundAttributes,
  /**
   * Simpleattributes are ones where the value contained in the attribute is
   * a simple type like number or string.
   */
  simple as Simple,
  /**
   * Computed attributes are attributes whose value is computed from one or more
   * other attributes.  The function given here as the "computation" must be a pure
   * function of its inputs.  It is not recommended that you even use constants in
   * this computation--factor them out to the callers instead.  (You can also create
   * constants that are attributes, if you want!)  The parameters to the function
   * are the _unwrapped_ values contained in the list of inputs, the second parameter.
   * They need to match in number and type or you will get an error.
   * For example, if your computation function expects (a:number, b:string) as
   * parameters, the list of inputs should have the types [Attribute<number>,
   * Attribute<string>].  When any member of the list of inputs changes value,
   * this function will be called.
   */
  computed as Computed,
  /**
   * Array attributes can cause updates of dependent attributes based on two
   * types of changes: 1) changes to the number of elements in the array and
   * 2) changes to attributes in the models that compose the array.
   * Every element of the array has to be Record and this Record (usually a model)
   * may include any number of attributes, changes to any of which will
   * cause updates to dependent attributes.
   * all your attribute are belong to us.
   *
   * This is not exported as Array because that conflicts with a type of the
   * same name in the default typescript library.
   */
  array as ArrayAttr,
  /**
   * naivearray is dumber than array.  it assumes that the parameter type T
   * is a simple type (not containing attributes) and that you only want
   * updates when the _number of values_ in the array changes.  This can
   * be useful if you are computing something based on the size of the array,
   * especially something like "is it empty?"
   */
  naivearray as NaiveArray,
  /**
   * Bind should be called typically in the interactor's constructor or
   * any other method that creates attributes for an interactor.
   * This informs evalvite that you want the changes to the model to force
   * redraws of the component.
   */
  bind as Bind,
  /**
   * Unbind is the reverse of Bind and probably
   * should be called when component is withdrawn from the screen.  If you don't unbind interactors
   * from their attributes, you can end up with wasted work as evalvite tries to
   * recompute vales for a interactor that can't use them.
   */
  unbind as Unbind,
  /**
   * Attribute is the base type of all attributes, but typically you want to
   * use one of the convenience functions to get a subclass, such as
   * Computed.
   */
  attribute as Attribute,
  /**
   * MinMax clamps the values of another attribute to a range of values.
   */
  minMax as MinMax,
  /**
   * Equal sets one number attribute to always be equal to another.
   */
  equal as Equal,
  /**
   * Updateable is the interface that "outside" code should use to get
   * notified about value changes of attributes.  Typically, this going to be
   * an interactor, but it does not have to be.
   */
  updateable as Updateable,
  /**
   * DecodeValue returns a structure that is the same as the input but with
   * attributes replaced with their values.
   */
  decodeValue as DecodeValue,
  /**
   * DecodeTypename returns a debugging structure that is parallel to the input
   * but with each value replaced by its typename, with attributes handled correctly.
   */
  decodeTypename as DecodeTypename,
};
