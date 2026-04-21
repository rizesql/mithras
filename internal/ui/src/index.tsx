import { render } from "solid-js/web";
import { Router, Route } from "@solidjs/router";
import "solid-devtools";
import "./style.css";
import "@fontsource-variable/geist-mono/wght.css";
import "@fontsource-variable/geist/wght.css";

import { Login } from "./pages/login";
import { Register } from "./pages/register";
import { ForgotPassword } from "./pages/forgot-password";
import { ResetPassword } from "./pages/reset-password";
import { Layout } from "./layout";

const root = document.getElementById("root");

if (import.meta.env.DEV && !(root instanceof HTMLElement)) {
  throw new Error(
    "Root element not found. Did you forget to add it to your index.html? Or maybe the id attribute got misspelled?",
  );
}

render(
  () => (
    <Router root={Layout}>
      <Route path="/login" component={Login} />
      <Route path="/register" component={Register} />
      <Route path="/forgot-password" component={ForgotPassword} />
      <Route path="/reset-password" component={ResetPassword} />
      <Route path="*" component={Login} />
    </Router>
  ),
  root!,
);
