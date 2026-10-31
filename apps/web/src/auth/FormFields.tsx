// Labeled input fields, shared by the login and sign-up forms.

type FieldProps = {
  label: string;
  value: string;
  onChange: (value: string) => void;
  // Tells the browser and password managers what the field is for, so they
  // can fill it in: "username", "current-password" or "new-password".
  autoComplete: string;
};

export function TextField(props: FieldProps) {
  return <Field {...props} type="text" />;
}

export function PasswordField(props: FieldProps) {
  // type="password" hides what is typed.
  return <Field {...props} type="password" />;
}

type InputProps = FieldProps & {
  type: "text" | "password";
};

// Wrapping the input in its <label> connects the two: clicking the label
// focuses the input, and screen readers read the label for the input.
function Field({ label, value, onChange, autoComplete, type }: InputProps) {
  return (
    <label className="flex flex-col gap-1 text-sm text-slate-700">
      {label}
      <input
        type={type}
        value={value}
        onChange={(event) => onChange(event.target.value)}
        autoComplete={autoComplete}
        required
        className="rounded border border-slate-300 px-3 py-2 text-base text-slate-900"
      />
    </label>
  );
}
