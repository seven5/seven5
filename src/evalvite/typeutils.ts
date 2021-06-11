import attribute, { updateable } from './attribute';
import attrPrivate from './base';

// without disabling this, we can't get the types right on obj (unknown won't work)
// eslint-disable-next-line @typescript-eslint/no-explicit-any,@typescript-eslint/explicit-module-boundary-types
export function instanceOfAttr(obj: any): obj is attrPrivate<any> {
  if (obj === null) {
    return false;
  }
  if (typeof obj !== 'object') {
    return false;
  }

  return 'addOutgoing' in obj && 'addIncoming' in obj && 'AddUpdateable' in obj;
}

// given a model, compute the field names that are attributes
// themselves.
export function modelToAttrFields(
  inst: Record<string, unknown>
): Array<string> {
  const result = new Array<string>();
  const objectKeys = Object.keys(inst) as Array<string>;
  objectKeys.forEach((k: string) => {
    const prop = inst[k];
    if (instanceOfAttr(prop)) {
      result.push(k);
    }
  });
  return result;
}

export function bind<T>(a: attribute<T>, u: updateable<T>): void {
  a.AddUpdateable(u);
}

export function unbind<T>(a: attribute<T>, u: updateable<T>): void {
  a.RemoveUpdateable(u);
}

export function bindModel(
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  inst: Record<string, any>,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  i: updateable<any>
): void {
  if (typeof inst !== 'object') {
    // protected from common error
    throw new Error(
      'models must be objects with exactly one level, e.g. {name:"mr. foo", address:"123 fleazil st"}'
    );
  }
  const fields = modelToAttrFields(inst);
  fields.forEach((k) => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (inst[k] as attrPrivate<any>).AddUpdateable(i);
  });
}
