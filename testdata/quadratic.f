      PRINT 1
 1    FORMAT(14HTYPE A, B, C: $)
      READ 2, A, B, C
 2    FORMAT(3F6.2)
      D = B**2 - 4*A*C
      IF (D) 10, 20, 30
 10   PRINT 11
 11   FORMAT(13HNO REAL ROOTS)
      GO TO 50
 20   X = -B / (2*A)
      PRINT 21, X
 21   FORMAT(4HX = , F6.2)
      GO TO 50
 30   X1 = (-B + SQRT(D)) / (2 * A)
      X2 = (-B - SQRT(D)) / (2 * A)
      PRINT 31, X1, X2
 31   FORMAT(5HX1 = , F6.2, 2H, , 5HX2 = , F6.2)
      GO TO 50
 50   STOP
      END