import 'package:flutter/material.dart';

/// Custom text field widget
class CustomTextField extends StatefulWidget {
  final TextEditingController controller;
  final String label;
  final String? hintText;
  final bool obscureText;
  final TextInputType keyboardType;
  final String? Function(String?)? validator;
  final Widget? prefixIcon;
  final Widget? suffixIcon;
  final bool enabled;
  final int maxLines;
  final TextCapitalization textCapitalization;
  final void Function(String)? onChanged;

  const CustomTextField({
    super.key,
    required this.controller,
    required this.label,
    this.hintText,
    this.obscureText = false,
    this.keyboardType = TextInputType.text,
    this.validator,
    this.prefixIcon,
    this.suffixIcon,
    this.enabled = true,
    this.maxLines = 1,
    this.textCapitalization = TextCapitalization.none,
    this.onChanged,
  });

  @override
  State<CustomTextField> createState() => _CustomTextFieldState();
}

class _CustomTextFieldState extends State<CustomTextField> {
  bool _obscureText = false;

  @override
  void initState() {
    super.initState();
    _obscureText = widget.obscureText;
  }

  @override
  Widget build(BuildContext context) {
    return TextFormField(
      controller: widget.controller,
      obscureText: _obscureText,
      keyboardType: widget.keyboardType,
      validator: widget.validator,
      enabled: widget.enabled,
      maxLines: widget.maxLines,
      textCapitalization: widget.textCapitalization,
      onChanged: widget.onChanged,
      decoration: InputDecoration(
        labelText: widget.label,
        hintText: widget.hintText,
        prefixIcon: widget.prefixIcon,
        suffixIcon: widget.obscureText
            ? IconButton(
                icon: Icon(
                  _obscureText ? Icons.visibility : Icons.visibility_off,
                ),
                onPressed: () {
                  setState(() {
                    _obscureText = !_obscureText;
                  });
                },
              )
            : widget.suffixIcon,
      ),
    );
  }
}

/// Email text field
class EmailTextField extends StatelessWidget {
  final TextEditingController controller;
  final String? Function(String?)? validator;

  const EmailTextField({
    super.key,
    required this.controller,
    this.validator,
  });

  @override
  Widget build(BuildContext context) {
    return CustomTextField(
      controller: controller,
      label: 'Email',
      hintText: 'Enter your email',
      keyboardType: TextInputType.emailAddress,
      prefixIcon: const Icon(Icons.email_outlined),
      validator: validator,
      textCapitalization: TextCapitalization.none,
    );
  }
}

/// Password text field
class PasswordTextField extends StatelessWidget {
  final TextEditingController controller;
  final String? Function(String?)? validator;
  final String label;

  const PasswordTextField({
    super.key,
    required this.controller,
    this.validator,
    this.label = 'Password',
  });

  @override
  Widget build(BuildContext context) {
    return CustomTextField(
      controller: controller,
      label: label,
      hintText: 'Enter your password',
      obscureText: true,
      prefixIcon: const Icon(Icons.lock_outline),
      validator: validator,
    );
  }
}

/// Username text field
class UsernameTextField extends StatelessWidget {
  final TextEditingController controller;
  final String? Function(String?)? validator;

  const UsernameTextField({
    super.key,
    required this.controller,
    this.validator,
  });

  @override
  Widget build(BuildContext context) {
    return CustomTextField(
      controller: controller,
      label: 'Username',
      hintText: 'Choose a username',
      prefixIcon: const Icon(Icons.person_outline),
      validator: validator,
      textCapitalization: TextCapitalization.none,
    );
  }
}
