#!/bin/bash

# Author: Daniel Lawrence (@wavemechanics)
# Sed script to refactor log.go import and logging statements in .go files to match new syntax.
# Note: the log.Error(x) -> log.FormatErrors([]error{x})
# mapping is slightly fragile.
# If "x" has nested parens or anything complicated, it may
# not work.
# Seems to work on the cases I have run across, but sometimes
# this is only by accident.
# I figure special cases can be fixed by hand.

sed -E -e '
:top
    # Skip any substitutions on comment lines
    #
    /^[[:blank:]]*\/\//{
        n
        b top
    }


    # FATAL events are mapped like this:
    #
    # log.Event(ctx, string, log.FATAL, log.Error(something1), something2) -> log.Fatal(ctx, string, something1, something2)
    #
    /log\.Event.*log\.FATAL/{

        # change the function name
        s/log\.Event/log.Fatal/

        # remove the severity
        s/,[[:blank:]]*log.FATAL//

        # extract the log.Error argument
        s/log.Error\(([^)]*)\)/\1/

        # and go to the next line
        n
        b top
    }

    # And ERROR events are similar:
    #
    # log.Event(ctx, string, log.ERROR log.Error(something1), something2) -> log.Error(ctx, string, something1, something2)
    #
    /log\.Event.*log\.ERROR/{

        # remove the severity
        s/,[[:blank:]]*log.ERROR//

        # extract the log.Error argument
        s/log.Error\(([^)]*)\)/\1/

        # change the function name
        s/log\.Event/log.Error/

        # and go to the next line
        n
        b top
    }

    # Info events are mapped like this:
    #
    # log.Event(ctx, string, log.INFO, log.Error(something)) -> log.Info(ctx, string, log.FormatErrors([]error{something}))
    #
    /log\.Event.*log.INFO/{

        # remove the severity
        s/,[[:blank:]]*log.INFO//

        # Change .Error to .FormatErrors
        s/log\.Error\(([^)]*)\)/log.FormatErrors([]error{\1})/

        # change the function name
        s/log\.Event/log.Info/

        # and go to the next line
        n
        b top
    }

    # Warn events are mapped like info:
    #
    # log.Event(ctx, string, log.WARN, log.Error(something)) -> log.Warn(ctx, string, log.FormatErrors([]error{something}))
    #
    /log\.Event.*log.WARN/{

        # remove the severity
        s/,[[:blank:]]*log.WARN//

        # Change .Error to .FormatErrors
        s/log\.Error\(([^)]*)\)/log.FormatErrors([]error{\1})/

        # change the function name
        s/log\.Event/log.Warn/

        # and go to the next line
        n
        b top
    }

    # outdated log module version imports are mapped like this:
    #
    # "github.com/ONSdigital/log.go/log" -> "github.com/ONSdigital/log.go/v2/log"
    # "github.com/ONSdigital/log.go/v1/log" -> "github.com/ONSdigital/log.go/v2/log"
    #
    /"github.com\/ONSdigital\/log.go.*\/log"/{

        # Change whatever is between log.go and /log with /v2
        s/log.go.*\/log/log.go\/v2\/log/

        # and go to the next line
        n
        b top
    }
'
